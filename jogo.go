package main

import (
    "bufio"
    "github.com/nsf/termbox-go"
    "os"
    "fmt"
)

// Define os elementos do jogo
type Elemento struct {
    simbolo rune
    cor termbox.Attribute
    corFundo termbox.Attribute
    tangivel bool
}

// Personagem controlado pelo jogador
var personagem = Elemento{
    simbolo: '☺',
    cor: termbox.ColorBlack,
    corFundo: termbox.ColorDefault,
    tangivel: true,
}

// Parede
var parede = Elemento{
    simbolo: '▤',
    cor: termbox.ColorBlack|termbox.AttrBold|termbox.AttrDim,
    corFundo: termbox.ColorDarkGray,
    tangivel: true,
}

// Barrreira
var barreira = Elemento{
    simbolo: '#',
    cor: termbox.ColorRed,
    corFundo: termbox.ColorDefault,
    tangivel: true,
}

// Vegetação
var vegetacao = Elemento{
    simbolo: '♣',
    cor: termbox.ColorGreen,
    corFundo: termbox.ColorDefault,
    tangivel: false,
}

// Elemento vazio
var vazio = Elemento{
    simbolo: ' ',
    cor: termbox.ColorDefault,
    corFundo: termbox.ColorDefault,
    tangivel: false,
}

// Elemento para representar áreas não reveladas (efeito de neblina)
var neblina = Elemento{
    simbolo: '.',
    cor: termbox.ColorDefault,
    corFundo: termbox.ColorYellow,
    tangivel: false,
}

var mapa [][]Elemento
var posX, posY int
var ultimoElementoSobPersonagem = vazio
var statusMsg string

var efeitoNeblina = true
var revelado [][]bool
var raioVisao int = 3

func main() {
    err := termbox.Init()
    if err != nil {
        panic(err)
    }
    defer termbox.Close()

    carregarMapa("mapa.txt")
    if efeitoNeblina {
        revelarArea()
    }
    desenhaTudo()

    for {
        switch ev := termbox.PollEvent(); ev.Type {
        case termbox.EventKey:
            if ev.Key == termbox.KeyEsc {
                return // Sair do programa
            }
            if ev.Ch == 'e' {
                interagir()
            } else {
                mover(ev.Ch)
                if efeitoNeblina {
                    revelarArea()
                }
            }
            desenhaTudo()
        }
    }
}

func carregarMapa(nomeArquivo string) {
    arquivo, err := os.Open(nomeArquivo)
    if err != nil {
        panic(err)
    }
    defer arquivo.Close()

    scanner := bufio.NewScanner(arquivo)
    y := 0
    for scanner.Scan() {
        linhaTexto := scanner.Text()
        var linhaElementos []Elemento
        var linhaRevelada []bool
        for x, char := range linhaTexto {
            elementoAtual := vazio
            switch char {
            case parede.simbolo:
                elementoAtual = parede
            case barreira.simbolo:
                elementoAtual = barreira
            case vegetacao.simbolo:
                elementoAtual = vegetacao
            case personagem.simbolo:
                // Atualiza a posição inicial do personagem
                posX, posY = x, y
                elementoAtual = vazio
            }
            linhaElementos = append(linhaElementos, elementoAtual)
            linhaRevelada = append(linhaRevelada, false)
        }
        mapa = append(mapa, linhaElementos)
        revelado = append(revelado, linhaRevelada)
        y++
    }
    if err := scanner.Err(); err != nil {
        panic(err)
    }
}

func desenhaTudo() {
    termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
    for y, linha := range mapa {
        for x, elem := range linha {
            if efeitoNeblina == false || revelado[y][x] {
                termbox.SetCell(x, y, elem.simbolo, elem.cor, elem.corFundo)
            } else {
                termbox.SetCell(x, y, neblina.simbolo, neblina.cor, neblina.corFundo)
            }
        }
    }

    desenhaBarraDeStatus()

    termbox.Flush()
}

func desenhaBarraDeStatus() {
    for i, c := range statusMsg {
        termbox.SetCell(i, len(mapa)+1, c, termbox.ColorBlack, termbox.ColorDefault)
    }
    msg := "Use WASD para mover e E para interagir. ESC para sair."
    for i, c := range msg {
        termbox.SetCell(i, len(mapa)+3, c, termbox.ColorBlack, termbox.ColorDefault)
    }
}

func revelarArea() {
    minX := max(0, posX-raioVisao)
    maxX := min(len(mapa[0])-1, posX+raioVisao)
    minY := max(0, posY-raioVisao/2)
    maxY := min(len(mapa)-1, posY+raioVisao/2)

    for y := minY; y <= maxY; y++ {
        for x := minX; x <= maxX; x++ {
            // Revela as células dentro do quadrado de visão
            revelado[y][x] = true
        }
    }
}

func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}

func mover(comando rune) {
    dx, dy := 0, 0
    switch comando {
    case 'w':
        dy = -1
    case 'a':
        dx = -1
    case 's':
        dy = 1
    case 'd':
        dx = 1
    }
    novaPosX, novaPosY := posX+dx, posY+dy
    if novaPosY >= 0 && novaPosY < len(mapa) && novaPosX >= 0 && novaPosX < len(mapa[novaPosY]) &&
        mapa[novaPosY][novaPosX].tangivel == false {
        mapa[posY][posX] = ultimoElementoSobPersonagem // Restaura o elemento anterior
        ultimoElementoSobPersonagem = mapa[novaPosY][novaPosX] // Atualiza o elemento sob o personagem
        posX, posY = novaPosX, novaPosY // Move o personagem
        mapa[posY][posX] = personagem // Coloca o personagem na nova posição
    }
}

func interagir() {
    statusMsg = fmt.Sprintf("Interagindo em (%d, %d)", posX, posY)
}
