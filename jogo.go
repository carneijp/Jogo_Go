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
    cor: termbox.ColorRed,
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
// TODO: A partir daqui estamos criando novos bonecos.
// Boneco para matar personagem caso ele toque
var boneco = Elemento{
    simbolo: '☃',
    cor: termbox.ColorRed,
    corFundo: termbox.ColorDefault,
    tangivel: false,
}

var clockDor = Elemento{
    simbolo: '⏲',
    cor: termbox.ColorYellow,
    corFundo: termbox.ColorDefault,
    tangivel: false,
}

// Mapaa parece ser a variavel que representa uma matriz de elementos
var mapa [][]Elemento
// Entender o que esses posX, posY fazem
var posX, posY int
// Guarda o elemento onde o personagem estava antes
var ultimoElementoSobPersonagem = vazio
// Mensagem a ser mostrada na tela
var statusMsg string

// Define se vai aparecer a neblina
var efeitoNeblina = false
// matriz de mesmo tamanho do mapa, mostrando, onde pode ser mostrado o mapa real
var revelado [][]bool
// Defini o quanto de espaço em volta que ele vai abrir de visibilidade
var raioVisao int = 3

// Para proteger que o usuário não poderá mais mover depois de sofre gameover
var gameOver bool = false

// Para o usuário não poder mais mover depois de ganhar
var victory bool = false

func main() {
    err := termbox.Init()
    if err != nil {
        panic(err)
    }
    defer termbox.Close()
    // processa o mapa pré dezenhado no arquivo de texto
    carregarMapa("mapa.txt")
    // Caso ele deva colocar o efeito de neblina
    if efeitoNeblina {
        revelarArea()
    }
    // Manda recarregar a 'tela'
    desenhaTudo()
    
    // fica em looping procurando por comandos no teclado
    for {
        // Caso seja game over
        if gameOver {
            showGameOver()
        } else if victory{ // Caso o jogador tenha ganhado.

        } 
        
        switch ev := termbox.PollEvent(); ev.Type {
        case termbox.EventKey:
            // Verifica se a tecla chamada é para interromper o programa
            if ev.Key == termbox.KeyEsc {
                return // Sair do programa
            }
            
            if ev.Ch == 'e' { // caso não seja como ele vai operar.
                interagir()
            } else {
                mover(ev.Ch)
                // Após se mover, verifica se o efeito de neblina está ativo,
                // caso esteja manda revelar mais espaços descobertos
                if efeitoNeblina {
                    revelarArea()
                }
            }
            // Para cada comando de tecla, após processar o comando, manda recarregar 
            desenhaTudo()
        }
    }
}

// Funçao de carregar o mapa
func carregarMapa(nomeArquivo string) {
    // Aqui abre o arquivo.
    arquivo, err := os.Open(nomeArquivo)
    if err != nil {
        panic(err)
    }
    defer arquivo.Close()
    // Faz a leitura do arquivo
    scanner := bufio.NewScanner(arquivo)
    y := 0
    // Aqui passa em cada linha do arquivo mapa construindo ele na memoria.
    for scanner.Scan() {
        linhaTexto := scanner.Text()
        var linhaElementos []Elemento
        var linhaRevelada []bool
        // aqui passa em cada caracter da linha e adicona na sublinha
        for x, char := range linhaTexto {
            elementoAtual := vazio
            // Quando for adicionar novos elementos precisa adicionar aqui no switch
            switch char {
            case parede.simbolo:
                elementoAtual = parede
            case barreira.simbolo:
                elementoAtual = barreira
            case vegetacao.simbolo:
                elementoAtual = vegetacao
            case boneco.simbolo:
                elementoAtual = boneco
            case clockDor.simbolo:
                elementoAtual = clockDor
            case personagem.simbolo:
                // Atualiza a posição inicial do personagem
                posX, posY = x, y
                elementoAtual = vazio
            }
            linhaElementos = append(linhaElementos, elementoAtual)
            linhaRevelada = append(linhaRevelada, false)
        }
        // Coloca no mapa a nova linha.
        mapa = append(mapa, linhaElementos)
        revelado = append(revelado, linhaRevelada)
        y++
    }
    if err := scanner.Err(); err != nil {
        panic(err)
    }
}

// Funçao que manda reconstruir a tela com seu novo estado
func desenhaTudo() {
    // Manda limpar a tela antes de reconstruir a nova
    termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
    // passa por cada linha do mapa
    for y, linha := range mapa {
        // passa por cada elemento da linha do mapa
        for x, elem := range linha {
            // Verifica se o efeito de neblina está desativado
            if efeitoNeblina == false || revelado[y][x] {
                // Coloca o elemento correto para aparecer
                termbox.SetCell(x, y, elem.simbolo, elem.cor, elem.corFundo)
            } else {
                termbox.SetCell(x, y, neblina.simbolo, neblina.cor, neblina.corFundo)
            }
        }
    }
    // manda reescrever as informações de status
    desenhaBarraDeStatus()
    // manda resetar o terminal com as novas informações construidas
    termbox.Flush()
}

func showGameOver() {
    termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

    msg := "Game Over, mais sorte da proxima vez...."
    for i, c := range msg {
        termbox.SetCell(i, 1, c, termbox.ColorBlack, termbox.ColorDefault)
    }

    msg = "User ESC para sair."
    for i, c := range msg {
        termbox.SetCell(i, 3, c, termbox.ColorBlack, termbox.ColorDefault)
    }

    termbox.Flush()
}

// Construindo a mensagem
func desenhaBarraDeStatus() {
    // passa por cada caracter da mensagem, e escreve ele como se fossem elementos
    for i, c := range statusMsg {
        termbox.SetCell(i, len(mapa)+1, c, termbox.ColorBlack, termbox.ColorDefault)
    }

    // escreve a mensagem toda, e define a posição dela para ser 3 abaixo do mapa... 
    // Porque 3 posições abaixo do mapa? Eu posso somente reescrever as mensagens? Preciso reescrever o mapa todo?
    msg := "Use WASD para mover e E para interagir. ESC para sair."
    for i, c := range msg {
        termbox.SetCell(i, len(mapa)+3, c, termbox.ColorBlack, termbox.ColorDefault)
    }
    
}

// Função para revelar o mapa, a principio não utilizado
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

// Função que retorna qual o maior numero entre 2
func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}

// Função que retorna qual o menor numero entre 2
func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}

// Função para mover, ele recebe como argumento uma rune, O que é uma rune?
func mover(comando rune) {
    dx, dy := 0, 0
    // De todas as possiveis reclas, procuramos qual foi clicada
    switch comando {
    case 'w':
        // Podemos configurar para quando o usuário utilizar o poder do pular ele consegue dar mais de 1 passo por vez
        dy = -1
    case 'a':
        // Podemos configurar para quando o usuário utilizar o poder do pular ele consegue dar mais de 1 passo por vez
        dx = -1
    case 's':
        // Podemos configurar para quando o usuário utilizar o poder do pular ele consegue dar mais de 1 passo por vez
        dy = 1
    case 'd':
        // Podemos configurar para quando o usuário utilizar o poder do pular ele consegue dar mais de 1 passo por vez
        dx = 1
    }

    novaPosX, novaPosY := posX+dx, posY+dy
    if novaPosY >= 0 && novaPosY < len(mapa) && novaPosX >= 0 && novaPosX < len(mapa[novaPosY]) &&
        mapa[novaPosY][novaPosX].tangivel == false {
        // Coloca de volta no mapa o elemento em que o personagem esta ocupando o espaço antes
        mapa[posY][posX] = ultimoElementoSobPersonagem // Restaura o elemento anterior
        // salva na variavel momentanea o novo elemento que será retornado quando ele for movido
        ultimoElementoSobPersonagem = mapa[novaPosY][novaPosX] // Atualiza o elemento sob o personagem
        if ultimoElementoSobPersonagem == boneco {
            showGameOver()
            gameOver = true
        }
        // atualiza a posição nova do personagem
        posX, posY = novaPosX, novaPosY // Move o personagem
        mapa[posY][posX] = personagem // Coloca o personagem na nova posição
    }
}

// Função de interagir do personagem.
func interagir() {
    statusMsg = fmt.Sprintf("Interagindo em (%d, %d)", posX, posY)   
}
