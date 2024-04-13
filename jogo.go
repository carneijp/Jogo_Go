package main

import (
    "bufio"
    "github.com/nsf/termbox-go"
    "os"
    "fmt"
    "time"
    "sync"
)

// Define os elementos do jogo
type Elemento struct {
    simbolo rune
    cor termbox.Attribute
    corFundo termbox.Attribute
    tangivel bool
    posX int
    posY int
    alive bool
}

// Personagem controlado pelo jogador
var personagem = Elemento{
    simbolo: '🏃',
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
    posX: 0,
    posY: 0,
    alive: true,
}

var clockDor = Elemento{
    simbolo: '⏲',
    cor: termbox.ColorYellow,
    corFundo: termbox.ColorDefault,
    tangivel: false,
}

var jackPot = Elemento{
    simbolo: '💸',
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

// Para debug da posição do boneco
var debusMsg string

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

//  Variavel para controle de acesso a zona critica
var mu sync.Mutex

// Contagem de bonecos matados
var killCount int = 0

// // timer
var timeStart = time.Now()

// timeElapsed := 0,0


func resetGame() {
    timeStart = time.Now()
    killCount = 0
    gameOver = false
    victory = false
    raioVisao = 3
    revelado = nil
    statusMsg = ""
    efeitoNeblina = false
    ultimoElementoSobPersonagem = vazio
    posX, posY = 0, 0
    mapa = nil
}

func main() {
    resetGame()
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
    
    // aqui agora poderia iniciar o cronometro e mandar ele ficar atualizando o mostra tudo colocando um texto indicando tempo passado
    go timeMonitor()

    // fica em looping procurando por comandos no teclado
    for {
        // Caso seja game over
        if (gameOver || victory) {
            showEndGame()
            switch ev := termbox.PollEvent(); ev.Type {
            case termbox.EventKey:
                if ev.Key == termbox.KeyEsc {
                    return // sair do programa
                }
                if ev.Ch == 'r' {
                    main()
                }
            }
        }else {
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
        x := 0
        // aqui passa em cada caracter da linha e adicona na sublinha
        for _, char := range linhaTexto {
            elementoAtual := vazio
            // Quando for adicionar novos elementos precisa adicionar aqui no switch
            switch char {
            case parede.simbolo:
                elementoAtual = parede
                elementoAtual.posX = y
                elementoAtual.posY = x
            case barreira.simbolo:
                elementoAtual = barreira
                elementoAtual.posX = y
                elementoAtual.posY = x
            case vegetacao.simbolo:
                elementoAtual = vegetacao
                elementoAtual.posX = y
                elementoAtual.posY = x
            case boneco.simbolo:
                elementoAtual = boneco
                elementoAtual.posX = y
                elementoAtual.posY = x
                go bonecoUpAndDown(elementoAtual, 1)
            case clockDor.simbolo:
                elementoAtual = clockDor
                elementoAtual.posX = y
                elementoAtual.posY = x
            case jackPot.simbolo:
                elementoAtual = jackPot
                elementoAtual.posX = y
                elementoAtual.posY = x
            case personagem.simbolo:
                // Atualiza a posição inicial do personagem
                posX, posY = x, y
                elementoAtual = vazio
            }
            linhaElementos = append(linhaElementos, elementoAtual)
            linhaRevelada = append(linhaRevelada, false)
            x++
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
    mu.Lock()
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
    mu.Unlock()
}

func showEndGame() {
    termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
    if gameOver {
        msg := "Game Over, mais sorte da proxima vez...."
        for i, c := range msg {
            termbox.SetCell(i, 1, c, termbox.ColorBlack, termbox.ColorDefault)
        }
    } else if victory {
        msg := "🥳 Parabéns!!!! Você chegou até o fim, que tal irmos mais uma vez tente bater o seu tempo."
        for i, c := range msg {
            termbox.SetCell(i, 1, c, termbox.ColorBlack, termbox.ColorDefault)
        }
    }
    

    msg := "User ESC para sair Ou clique no R para reiniciar a partida... Boa sorte!🫣"
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

    timeElapsed := time.Since(timeStart)

    gameInfo := fmt.Sprintf("Contagem de kills: %d timeElapsed: %t", killCount, timeElapsed)
    for i, c:= range gameInfo {
        termbox.SetCell(i, len(mapa)+5, c, termbox.ColorBlack, termbox.ColorDefault)
    }

    for i, c := range debusMsg {
        termbox.SetCell(i, len(mapa)+6, c, termbox.ColorBlack, termbox.ColorDefault)
    }
    
}

func timeMonitor() {
    if !(gameOver || victory) {
        desenhaTudo()
        time.Sleep(200 * time.Millisecond)
        timeMonitor()
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
        if ultimoElementoSobPersonagem.simbolo == boneco.simbolo {
            gameOver = true
            showEndGame()
        } else if ultimoElementoSobPersonagem.simbolo == jackPot.simbolo {
            victory = true
            showEndGame()
        }
        // atualiza a posição nova do personagem
        posX, posY = novaPosX, novaPosY // Move o personagem
        mapa[posY][posX] = personagem // Coloca o personagem na nova posição
    }
}

// Função de interagir do personagem.
func interagir() {
    statusMsg = fmt.Sprintf("Interagindo em x, y: (%d, %d)", posX, posY)
    go atirandoFoguinho(posY, posX, false)
}

// Primeiro desparo de thread quando ele interage ele dispara fogo
var fire = Elemento{
    simbolo:'🔥',
    cor: termbox.ColorYellow,
    corFundo: termbox.ColorDefault,
    tangivel: false,
}

// Para o boneco disparar fogo, o boneco lança fogo somente para a direita
func atirandoFoguinho(x int, y int, inserido bool) {
    proxPosX, proxPosY := x, y+1
    proximoElementoLocal := mapa[proxPosX][proxPosY]
    debusMsg = fmt.Sprintf("Fogo Anda para: (%d, %d)", proxPosX, proxPosY)
    if proximoElementoLocal.simbolo == vazio.simbolo {
        if inserido {
            mapa[x][y] = vazio
        }
        mapa[proxPosX][proxPosY] = fire
        desenhaTudo()
        time.Sleep(200 * time.Millisecond)
        atirandoFoguinho(proxPosX, proxPosY, true)
    } else if proximoElementoLocal.simbolo == boneco.simbolo {
        if inserido {
            mapa[x][y] = vazio
        }
        mapa[proxPosX][proxPosY] = vazio
        proximoElementoLocal.alive = false
        mu.Lock()
        killCount++
        mu.Unlock()
        desenhaTudo()
    } else {
        if inserido {
            mapa[x][y] = vazio
        }
        desenhaTudo()
    }
}

// Para o boneco se movimentar para cima e para baixo recebe o boneco como argumento
func bonecoUpAndDown(boneco Elemento, direction int) {
    continuar := true
    time.Sleep(1 * time.Second)
    proxPosX, proxPosY := boneco.posX + direction , boneco.posY
    proximoElementoLocal := mapa[proxPosX][proxPosY]
    if proximoElementoLocal.simbolo == vazio.simbolo {
        // Vou me mover na direção inicial pois tenho espaço vazio a "frente"
        mapa[proxPosX][proxPosY] = boneco
        mapa[boneco.posX][boneco.posY] = vazio
        // Atualizando na memoria a nova posição do boneco
        boneco.posX = proxPosX
        boneco.posY = proxPosY
    } else if proximoElementoLocal.simbolo == personagem.simbolo {
        gameOver = true
        showEndGame()
    } else if proximoElementoLocal.simbolo == fire.simbolo {
        mapa[boneco.posX][boneco.posY] = vazio
        mapa[proxPosX][proxPosY] = vazio
        continuar = false
        killCount ++
    } else {
        direction *= -1
    }

    if (!(gameOver || victory) && continuar) {
        desenhaTudo()
        bonecoUpAndDown(boneco, direction)
    }
}