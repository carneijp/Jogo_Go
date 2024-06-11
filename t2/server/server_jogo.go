package main 

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"bufio"
	"fmt"
	"os"
	"sync"
	"time"
	"math/rand"
	"github.com/nsf/termbox-go"
	"github.com/google/uuid"
)

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
	IdOwner string
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


func (server *Server) TimeStarted(args *string, reply *time.Time) error {
	*reply = server.timeStart
	return nil
}

func (server *Server) RegisterNewPlayer(args *string, reply *PlayerRegistered) error {
	// Defini o id unico do player que está entrando
	var id = uuid.NewString()
	// Pega uma copia de um personagem para ser esse player
	newPlayer := personagem
	// Fixa o id para o id unico gerado para o player
	newPlayer.Id = id

	// Cria a entidade que guarda as informações sobre o player
	playerRegistered := PlayerRegistered{PosX: 0, PosY: 0, UltimoElementoSobPersonagem: vazio, KillCount: 0, Dead: false, Id: id, Personagem: newPlayer}
	playerRegistered.Id = id
	playerRegistered.UltimoElementoSobPersonagem = vazio 

	ultimoElemento := boneco
	randomIntX := 0
	randomIntY := 0

	// Procurando por uma posição vazia para colocar o player
	server.mapaMu.Lock()
	for ultimoElemento != vazio {
		// Buscando por uma posição aleatoria no mapa para colocar o player novo...
		minY := 0
		maxY := len(server.mapa) - 1
		randomIntY = rand.Intn(maxY-minY+1) + minY

		minX := 0
		maxX := len(server.mapa[0]) - 1
		randomIntX = rand.Intn(maxX-minX+1) + minX

		ultimoElemento = server.mapa[randomIntY][randomIntX]
	}
	server.mapa[randomIntY][randomIntX] = newPlayer
	server.mapaMu.Unlock()
	// Guardando a informação de onde eu inicializei o player
	playerRegistered.PosX = randomIntX
	playerRegistered.PosY = randomIntY
	server.playerMu.Lock()
	server.players = append(server.players, playerRegistered)
	server.playerMu.Unlock()
	*reply = playerRegistered
	return nil
}

func (server *Server) GetMap(args *MapCall, reply *MapResponse) error {
	var player PlayerRegistered
	server.playerMu.Lock()
	for _, p := range server.players {
		if p.Id == args.PlayerId {
			player = p
			break
		}
	}
	server.playerMu.Unlock()

	server.mapaMu.Lock()
	*reply = MapResponse{Mapa: server.mapa, PlayerInformation: player}
	server.mapaMu.Unlock()
	return nil
}

func (server *Server) Move(args *MoveCall, reply *MapResponse) error {
	mover(args.PlayerId, args.Comand, server)
	var player PlayerRegistered
	server.playerMu.Lock()
	for _, p := range server.players {
		if p.Id == args.PlayerId {
			player = p
			break
		}
	}
	server.playerMu.Unlock()

	server.mapaMu.Lock()
	*reply = MapResponse{Mapa: server.mapa, PlayerInformation: player}
	server.mapaMu.Unlock()
	return nil
}

func mover(playerId string, comando rune, server *Server) {
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
	
	server.playerMu.Lock()
	for index, p := range server.players {
		if p.Id == playerId {
			novaPosX, novaPosY := p.PosX+dx, p.PosY+dy
			server.mapaMu.Lock()
			if novaPosY >= 0 && novaPosY < len(server.mapa) && novaPosX >= 0 && novaPosX < len(server.mapa[novaPosY]) &&
				server.mapa[novaPosY][novaPosX].Tangivel == false {
				if server.mapa[novaPosY][novaPosX].Simbolo == fire.Simbolo {
					server.mapaMu.Unlock()
					server.playerMu.Unlock()
					return
				}
				// Coloca de volta no mapa o elemento em que o personagem esta ocupando o espaço antes
				server.mapa[p.PosY][p.PosX] = p.UltimoElementoSobPersonagem // Restaura o elemento anterior
				// salva na variavel momentanea o novo elemento que será retornado quando ele for movido
				p.UltimoElementoSobPersonagem = server.mapa[novaPosY][novaPosX] // Atualiza o elemento sob o personagem
				if p.UltimoElementoSobPersonagem.Simbolo == boneco.Simbolo {
					p.Dead = true
				}

				// atualiza a posição nova do personagem
				p.PosX, p.PosY = novaPosX, novaPosY // Move o personagem
				server.mapa[p.PosY][p.PosX] = p.Personagem   // Coloca o personagem na nova posição
			}
			server.mapaMu.Unlock()
			server.players[index] = p
			break
		}
	}
	server.playerMu.Unlock()
}

func (server *Server) Interagir(args *InteragirCall, reply *MapResponse) error {
	atirandoFoguinho(args.PosY, args.PosX, false, args.PlayerId, server)
	var player PlayerRegistered
	server.playerMu.Lock()
	for _, p := range server.players {
		if p.Id == args.PlayerId {
			player = p
			break
		}
	}
	server.playerMu.Unlock()

	server.mapaMu.Lock()
	*reply = MapResponse{Mapa: server.mapa, PlayerInformation: player}
	server.mapaMu.Unlock()
	return nil
}

func atirandoFoguinho(x int, y int, inserido bool, idOwner string, server *Server) {
	if inserido {
		time.Sleep(200 * time.Millisecond)
	}
	proxPosX, proxPosY := x, y+1
	server.mapaMu.Lock()
	proximoElementoLocal := server.mapa[proxPosX][proxPosY]
	// debusMsg = fmt.Sprintf("Fogo Anda para: (%d, %d)", proxPosX, proxPosY)
	if proximoElementoLocal.Simbolo == vazio.Simbolo {
		if inserido {
			server.mapa[x][y] = vazio
		}
		server.mapa[proxPosX][proxPosY] = fire
		server.mapa[proxPosX][proxPosY].IdOwner = idOwner
		go atirandoFoguinho(proxPosX, proxPosY, true, idOwner, server)
	} else if proximoElementoLocal.Simbolo == boneco.Simbolo {
		if inserido {
			server.mapa[x][y] = vazio
		}
		server.mapa[proxPosX][proxPosY] = vazio
		proximoElementoLocal.Alive = false
		// Aumentando a contagem de mortos para aquele player
		server.playerMu.Lock()
		for index, p := range server.players {
			if p.Id == idOwner {
				p.KillCount++
				server.players[index] = p
				break
			}
		}
		server.playerMu.Unlock()
		// Colocando o boneco na lista de mortos
		server.mortosMu.Lock()
		server.mortos = append(server.mortos, proximoElementoLocal.Id)
		server.mortosMu.Unlock()
	} else if proximoElementoLocal.Simbolo == personagem.Simbolo {
		if proximoElementoLocal.Id != idOwner {
			// primeiro aumenta a contagem de kill do player e mudar o player atingido para motor
			server.playerMu.Lock()
			for index, p := range server.players {
				if p.Id == idOwner {
					p.KillCount++
					server.players[index] = p
				} else if p.Id == proximoElementoLocal.Id {
					p.Dead = true
					server.players[index] = p
				}
			}
			server.playerMu.Unlock()
			// agora transformar a posição para vazio.
			if inserido {
				server.mapa[x][y] = vazio
			}
			server.mapa[proxPosX][proxPosY] = vazio
		} else {
			// só faz o fogo passar pelo player sem mudar o visual
			if inserido {
				server.mapa[x][y] = vazio
			}
			go atirandoFoguinho(proxPosX, proxPosY, true, idOwner, server)
		}
	} else {
		if inserido {
			server.mapa[x][y] = vazio
		}
		
	}
	server.mapaMu.Unlock()
}

func resetGame(server *Server) {
	server.mortos = nil
	server.players = nil
	server.efeitoNeblina = false
	server.timeStart = time.Now()
	server.mapa = nil
}

// Funçao de inicializar o jogo
func inicializar(nomeArquivo string, server *Server) {
	resetGame(server)
	// Aqui abre o arquivo.
	arquivo, err := os.Open(nomeArquivo)
	if err != nil {
		panic(err)
	}
	defer arquivo.Close()
	// Faz a leitura do arquivo
	scanner := bufio.NewScanner(arquivo)
	id := 0
	y := 0
	// Aqui passa em cada linha do arquivo mapa construindo ele na memoria.
	for scanner.Scan() {
		linhaTexto := scanner.Text()
		var linhaElementos []Elemento
		// var linhaRevelada []bool
		x := 0
		// aqui passa em cada caracter da linha e adicona na sublinha
		for _, char := range linhaTexto {
			elementoAtual := vazio
			// Quando for adicionar novos elementos precisa adicionar aqui no switch
			switch char {
				case parede.Simbolo:
					elementoAtual = parede
					elementoAtual.PosX = y
					elementoAtual.PosY = x
				case barreira.Simbolo:
					elementoAtual = barreira
					elementoAtual.PosX = y
					elementoAtual.PosY = x
				case vegetacao.Simbolo:
					elementoAtual = vegetacao
					elementoAtual.PosX = y
					elementoAtual.PosY = x
				case boneco.Simbolo:
					elementoAtual = boneco
					elementoAtual.PosX = y
					elementoAtual.PosY = x
					elementoAtual.Id = uuid.NewString()
					go bonecoUpAndDown(elementoAtual, 1, server)
				case clockDor.Simbolo:
					elementoAtual = clockDor
					elementoAtual.PosX = y
					elementoAtual.PosY = x
				case jackPot.Simbolo:
					elementoAtual = jackPot
					elementoAtual.PosX = y
					elementoAtual.PosY = x
				case personagem.Simbolo:
					fmt.Sprintf("Não deveria ter personagem para ser construido no mapa, nenhum player se registrou ainda...")
					// Atualiza a posição inicial do personagem
					// posX, posY = x, y
					// elementoAtual = vazio
			}
			linhaElementos = append(linhaElementos, elementoAtual)
			// linhaRevelada = append(linhaRevelada, false)
			x++
			id++
		}
		// Coloca no mapa a nova linha.
		server.mapa = append(server.mapa, linhaElementos)
		// revelado = append(revelado, linhaRevelada)
		y++
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
}

// Para o boneco se movimentar para cima e para baixo recebe o boneco como argumento
func bonecoUpAndDown(boneco Elemento, direction int, server *Server) {
	continuar := true
	time.Sleep(1 * time.Second)
	server.mapaMu.Lock()
	for _, i := range server.mortos {
		if i == boneco.Id {
			server.mapaMu.Unlock()
			return
		}
	}
	proxPosX, proxPosY := boneco.PosX+direction, boneco.PosY
	if (proxPosX >= len(server.mapa) || proxPosX < 0) {
		direction *= -1
		server.mapaMu.Unlock()
		bonecoUpAndDown(boneco, direction, server)
	} 
	proximoElementoLocal := server.mapa[proxPosX][proxPosY]
	if proximoElementoLocal.Simbolo == vazio.Simbolo {
		// Vou me mover na direção inicial pois tenho espaço vazio a "frente"
		server.mapa[proxPosX][proxPosY] = boneco
		server.mapa[boneco.PosX][boneco.PosY] = vazio
		// Atualizando na memoria a nova posição do boneco
		boneco.PosX = proxPosX
		boneco.PosY = proxPosY
	} else if proximoElementoLocal.Simbolo == personagem.Simbolo {
		// Procurando o player na memoria para colocar ele como sendo um player morto
		server.playerMu.Lock()
		for index, p := range server.players {
			if p.Id == proximoElementoLocal.Id {
				p.Dead = true
				server.players[index] = p
				break
			}
		}
		server.playerMu.Unlock()

		// Após matar o player vou me mover na direção da inercia
		server.mapa[proxPosX][proxPosY] = boneco
		server.mapa[boneco.PosX][boneco.PosY] = vazio
		// Atualizando a posição do boneco
		boneco.PosX = proxPosX
		boneco.PosY = proxPosY
		// server.deadPlayers = append(server.deadPlayers, proximoElementoLocal.id)
	} else if proximoElementoLocal.Simbolo == fire.Simbolo {
		idOwner := proximoElementoLocal.IdOwner
		server.playerMu.Lock()
		for index, p := range server.players {
			if p.Id == idOwner {
				p.KillCount++
				server.players[index] = p
				break
			}
		}
		server.playerMu.Unlock()
		// DO FOGO E NÃO O FOGO ANDA EM CIMA DO BONECO. PRECISA SER ARRUMADO
		server.mapa[boneco.PosX][boneco.PosY] = vazio
		server.mapa[proxPosX][proxPosY] = vazio
		continuar = false
		server.mortos = append(server.mortos, boneco.Id)
	} else {
		direction *= -1
	}
	server.mapaMu.Unlock()
	if continuar {
		bonecoUpAndDown(boneco, direction, server)
	}
}

// Fazer uma função para spawnar um novo boneco em um local aleatorio do mapa que tenha posição de casa vazia

func spawnerDeBoneco(server *Server) {
	contador := 0
	randomIntX := 0
	randomIntY := 0
	for true {
		ultimoElemento := boneco
		var id = uuid.NewString()
		newBoneco := boneco
		boneco.Id = id
		server.mapaMu.Lock()
		for ultimoElemento != vazio {
			// Buscando por uma posição aleatoria no mapa para colocar o player novo...
			minY := 0
			maxY := len(server.mapa) -1
			randomIntY = rand.Intn(maxY-minY+1) + minY
			minX := 0
			maxX := len(server.mapa[0]) -1

			randomIntX = rand.Intn(maxX-minX+1) + minX
			ultimoElemento = server.mapa[randomIntY][randomIntX]
			newBoneco.PosX = randomIntY
			newBoneco.PosY = randomIntX
		}
		server.mapa[randomIntY][randomIntX] = newBoneco
		server.mapaMu.Unlock()
		go bonecoUpAndDown(newBoneco, 1, server)
		
		time.Sleep(5 * time.Second)
		contador ++
		if (contador == 20) {
			go spawnerDeBoneco(server)
		}else if (contador == 100) {
			break
		}
	}
}

func main() {
	//Inicializar um objeto do tipo dos metodos exportaveis
	server := new(Server)
	inicializar("mapa.txt", server)
	go spawnerDeBoneco(server)
	/*
		Para que seja possivel acessar os metodos do objeto
		eh necessario registra-lo utilizando a biblioteca rpc.
		O registro gera um erro, sendo nil o caso em que o registro
		foi um sucesso.
	*/
	err := rpc.Register(server)
	if err != nil {
		log.Fatal("Error registering Server ", err)
	}

	//Permite que a biblioteca utilize http para comunicacao
	rpc.HandleHTTP()

	/*
		Inicializa um processo que escuta toda comunicacao em 
		determinada porta, seguindo o protocolo tcp
	*/
	listener, err := net.Listen("tcp", ":2403")
	if err != nil {
		log.Fatal("Listener error ", err)
	}
	log.Printf("Serving rpc on port: %d", 2403)

	/*
		Ativa o servidor na porta e com o protocolo definido
		pelo listener
	*/
	http.Serve(listener, nil)
	if err != nil {
		log.Fatal("Error serving: ", err)
	}
}