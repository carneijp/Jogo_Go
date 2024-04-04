package main

import (
	"bufio"
	"github.com/nsf/termbox-go"
	"os"
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

var mapa [][]Elemento
var posX, posY int
var ultimoElementoSobPersonagem = vazio

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	carregarMapa("mapa.txt")
	desenhaTudo()

	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			if ev.Key == termbox.KeyEsc {
				return // Sair do programa
			}
			mover(ev.Ch)
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
                // Atualiza a posição incial do personagem
                posX, posY = x, y
                elementoAtual = vazio
            }
            linhaElementos = append(linhaElementos, elementoAtual)
        }
        mapa = append(mapa, linhaElementos)
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
			termbox.SetCell(x, y, elem.simbolo, elem.cor, elem.corFundo)
		}
	}

	msg := "Use WASD para mover. Pressione ESC para sair."
	for i, c := range msg {
		termbox.SetCell(i, len(mapa)+2, c, termbox.ColorBlack, termbox.ColorDefault)
	}

	termbox.Flush()
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
