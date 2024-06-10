package main

import (
	// "bufio"
	"fmt"
	"net/rpc"
	"log"
	// "os"
	"sync"
	"time"
	"github.com/nsf/termbox-go"
)

// Define os elementos do jogo
// Sofrendo com problemas de import então preciso colocar a Declaração dos objetos em ambos os arquivos, tanto do jogo como o arquivo do Servidor


// Funções genericas presentes em ambos os arquivos, server and client
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


// Define os elementos do jogo
type Elemento struct {
	Id string
	KillCount int
	Simbolo  rune
	Cor      termbox.Attribute
	CorFundo termbox.Attribute
	Tangivel bool
	PosX     int
	PosY     int
	Alive    bool
}

// Personagem controlado pelo jogador
var personagem = Elemento{
	Simbolo:  '🏃',
	KillCount: 0,
	Cor:      termbox.ColorRed,
	CorFundo: termbox.ColorDefault,
	Tangivel: true,
}

// Parede
var parede = Elemento{
	Simbolo:  '▤',
	Cor:      termbox.ColorBlack | termbox.AttrBold | termbox.AttrDim,
	CorFundo: termbox.ColorDarkGray,
	Tangivel: true,
}

// Barrreira
var barreira = Elemento{
	Simbolo:  '#',
	Cor:      termbox.ColorRed,
	CorFundo: termbox.ColorDefault,
	Tangivel: true,
}

// Vegetação
var vegetacao = Elemento{
	Simbolo:  '♣',
	Cor:      termbox.ColorGreen,
	CorFundo: termbox.ColorDefault,
	Tangivel: false,
}

// Elemento vazio
var vazio = Elemento{
	Simbolo:  ' ',
	Cor:      termbox.ColorDefault,
	CorFundo: termbox.ColorDefault,
	Tangivel: false,
}

// Elemento para representar áreas não reveladas (efeito de neblina)
var neblina = Elemento{
	Simbolo:  '.',
	Cor:      termbox.ColorDefault,
	CorFundo: termbox.ColorYellow,
	Tangivel: false,
}

// Boneco para matar personagem caso ele toque
var boneco = Elemento{
	Simbolo:  '☃',
	Cor:      termbox.ColorRed,
	CorFundo: termbox.ColorDefault,
	Tangivel: false,
	PosX:     0,
	PosY:     0,
	Alive:    true,
}

var clockDor = Elemento{
	Simbolo:  '⏲',
	Cor:      termbox.ColorYellow,
	CorFundo: termbox.ColorDefault,
	Tangivel: false,
}

var jackPot = Elemento{
	Simbolo:  '💸',
	Cor:      termbox.ColorYellow,
	CorFundo: termbox.ColorDefault,
	Tangivel: false,
}

var fire = Elemento{
	Simbolo:  '🔥',
	Cor:      termbox.ColorYellow,
	CorFundo: termbox.ColorDefault,
	Tangivel: false,
}

// Objeto para registrar player assim o player sabe onde ele está no mapa.
type PlayerRegistered struct {
	PosX, PosY int
	Personagem Elemento
	UltimoElementoSobPersonagem Elemento
	KillCount int
	Dead bool
	Id string
}

// Objeto para registrar player assim o player sabe onde ele está no mapa.
type MapCall struct {
	PlayerId string
}

type MapResponse struct {
	PlayerInformation PlayerRegistered
	Mapa [][]Elemento
}

type MoveCall struct {
	PlayerId string
	Comand rune
}

type InteragirCall struct {
	PlayerId string
	PosX, PosY int
}

// Defini variaveis que o servidor guarda na memoria dele para analise das funções e envios...
type Server struct {
	// Mapa é uma variavel representada por uma matriz de elementos
	mapa [][]Elemento

	// Registra a lista de itens que foram deletados
	mortos []string

	// Guarda as informações sobre os Player como quantas kills fizeram...
	players []PlayerRegistered

	// Define se vai aparecer a neblina
	efeitoNeblina bool

	// Variavel para controle de acesso a zona critica
	mapaMu sync.Mutex

	// Variavel para controle de acesso a zona critica do array de players
	playerMu sync.Mutex

	// Varaiavel para controle de acesso a zona critica do array de bonecos mortos
	mortosMu sync.Mutex

	timeStart time.Time
}

// Mapaa parece ser a variavel que representa uma matriz de elementos
var mapa [][]Elemento
// Variavel para controle de acesso a zona critica do mapa
var mapaMu sync.Mutex

// Variavel para controle de acesso a zona critica relacionada ao player como por exemplo a posição a quantidade de kills etc
var playerMu sync.Mutex

// UUID do player
var playerId string

// Entender o que esses posX, posY fazem
var posX, posY int

// Guarda o elemento onde o personagem estava antes
var ultimoElementoSobPersonagem = vazio

// Guarda a contagem de kills do player
var killCount int

// Mensagem a ser mostrada na tela
var statusMsg string

// Para debug da posição do boneco
var debusMsg string

// Para proteger que o usuário não poderá mais mover depois de sofre gameover
var gameOver bool = false

// // timer
var timeStart time.Time

func main() {
	c, err := rpc.DialHTTP("tcp", "localhost:2403")
	if err != nil {
		log.Fatal("Dialing: ", err)
	}

	err = termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	var reply_time_started_server time.Time
	// Vou sincronizar o relogio com o horario de inicio do servidor
	err = c.Call("Server." + "TimeStarted", "", &reply_time_started_server)
	if err != nil {
		log.Fatal("Server error: ", err)
	} else {
		fmt.Printf("Horario de inicio do servidor armazendao com sucesso\n")
		timeStart = reply_time_started_server
	}

	var reply_register PlayerRegistered
	// Primeiro vou registrar a entrada do player no servidor, caso ele consiga, ele irá buscar o mapa.
	err = c.Call("Server." + "RegisterNewPlayer", "", &reply_register)
	if err != nil {
		log.Fatal("Server error: ", err)
	} else {
		playerMu.Lock()
		killCount = reply_register.KillCount
		posX = reply_register.PosX
		posY = reply_register.PosY
		playerId = reply_register.Id
		ultimoElementoSobPersonagem = reply_register.UltimoElementoSobPersonagem
		if (reply_register.Dead) {
			gameOver = true
			showEndGame()
		}
		fmt.Printf("Result = player registered\n")
		playerMu.Unlock()
	}


	// processa o mapa pré dezenhado no arquivo de texto, colocando aqui tambem a função de monitorar o cronometro
	go buscaMapa()

	// fica em looping procurando por comandos no teclado
	for {
		// Caso seja game over
		if gameOver {
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
		} else {
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
				}
				// Para cada comando de tecla, após processar o comando, manda recarregar
				desenhaTudo()
			}
		}
	}
}

// Funçao de carregar o mapa
func buscaMapa() {
	for true {
		if (gameOver == false) {
			// TODO: COLOCAR AQUI CHAMADA AO SERVIDOR PARA ELE CARREGAR O MAPA
			c, err := rpc.DialHTTP("tcp", "localhost:2403")
			if err != nil {
				log.Fatal("Dialing: ", err)
			}
			var reply_map MapResponse
			err = c.Call("Server." + "GetMap", MapCall{PlayerId: playerId}, &reply_map)
			
			if err != nil {
				log.Fatal("Server error: ", err)
			} else {
				// Atualiza estatisticas do player
				playerMu.Lock()
				killCount = reply_map.PlayerInformation.KillCount
				posX = reply_map.PlayerInformation.PosX
				posY = reply_map.PlayerInformation.PosY
				playerId = reply_map.PlayerInformation.Id
				ultimoElementoSobPersonagem = reply_map.PlayerInformation.UltimoElementoSobPersonagem
				if (reply_map.PlayerInformation.Dead) {
					gameOver = true
					showEndGame()
				}
				playerMu.Unlock()
				// Atualiza estatisticas do mapa
				mapaMu.Lock()
				mapa = reply_map.Mapa
				mapaMu.Unlock()
			}
			desenhaTudo()
		}
		time.Sleep(200 * time.Millisecond)
	}
}

// Funçao que manda reconstruir a tela com seu novo estado
func desenhaTudo() {
	mapaMu.Lock()
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	for y, linha := range mapa {
		// passa por cada elemento da linha do mapa
		for x, elem := range linha {
			// Coloca o elemento correto para aparecer
			termbox.SetCell(x, y, elem.Simbolo, elem.Cor, elem.CorFundo)
		}
	}
	// manda reescrever as informações de status
	desenhaBarraDeStatus()
	// manda resetar o terminal com as novas informações construidas
	termbox.Flush()
	mapaMu.Unlock()
}

func showEndGame() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	if gameOver {
		msg := "Game Over, mais sorte da proxima vez...."
		for i, c := range msg {
			termbox.SetCell(i, 1, c, termbox.ColorBlack, termbox.ColorDefault)
		}
	}

	msg := "User ESC para sair da partida 😪"
	for i, c := range msg {
		termbox.SetCell(i, 3, c, termbox.ColorBlack, termbox.ColorDefault)
	}

	termbox.Flush()
}

// Construindo a mensagem
func desenhaBarraDeStatus() {
	playerMu.Lock()
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
	//  converao do tempo para um valor em segundos
	timeElapsedSeconds := float64(timeElapsed) / float64(time.Second)
	gameInfo := fmt.Sprintf("Contagem de kills: %d Cronômetro: %.2f segundos", killCount, timeElapsedSeconds)

	for i, c := range gameInfo {
		termbox.SetCell(i, len(mapa)+5, c, termbox.ColorBlack, termbox.ColorDefault)
	}

	for i, c := range debusMsg {
		termbox.SetCell(i, len(mapa)+6, c, termbox.ColorBlack, termbox.ColorDefault)
	}
	playerMu.Unlock()
}

// Função para mover, ele recebe como argumento uma rune, O que é uma rune?
func mover(comando rune) {
	c, err := rpc.DialHTTP("tcp", "localhost:2403")
	if err != nil {
		log.Fatal("Dialing: ", err)
	}
	var reply_map MapResponse

	err = c.Call("Server." + "Move", MoveCall{PlayerId: playerId, Comand: comando}, &reply_map)
	if (err != nil) {
		log.Fatal("Server error: ", err)
	} else {
		playerMu.Lock()
		killCount = reply_map.PlayerInformation.KillCount
		posX = reply_map.PlayerInformation.PosX
		posY = reply_map.PlayerInformation.PosY
		playerId = reply_map.PlayerInformation.Id
		ultimoElementoSobPersonagem = reply_map.PlayerInformation.UltimoElementoSobPersonagem
		if (reply_map.PlayerInformation.Dead) {
			gameOver = true
			showEndGame()
		}
		playerMu.Unlock()
		mapaMu.Lock()
		mapa = reply_map.Mapa
		mapaMu.Unlock()
		desenhaTudo()
	}
}

// Função de interagir do personagem.
func interagir() {
	c, err := rpc.DialHTTP("tcp", "localhost:2403")
	if err != nil {
		log.Fatal("Dialing: ", err)
	}
	var reply_map MapResponse

	err = c.Call("Server." + "Interagir", InteragirCall{PlayerId: playerId, PosX: posX, PosY: posY}, &reply_map)
	if (err != nil) {
		log.Fatal("Server error: ", err)
	} else {
		playerMu.Lock()
		killCount = reply_map.PlayerInformation.KillCount
		posX = reply_map.PlayerInformation.PosX
		posY = reply_map.PlayerInformation.PosY
		playerId = reply_map.PlayerInformation.Id
		ultimoElementoSobPersonagem = reply_map.PlayerInformation.UltimoElementoSobPersonagem
		if (reply_map.PlayerInformation.Dead) {
			gameOver = true
			showEndGame()
		}
		statusMsg = fmt.Sprintf("Interagindo em x, y: (%d, %d)", posX, posY)
		playerMu.Unlock()
		mapaMu.Lock()
		mapa = reply_map.Mapa
		mapaMu.Unlock()
		desenhaTudo()
	}
}