package utils

import (
	"sync"
)

// SortRequest rappresenta la richiesta del client, che include i dati da ordinare.
type SortRequest struct {
	Data          []int   // Array di dati da ordinare
	WorkerResults [][]int // Risultati ordinati dai worker
}

// ClientResponse rappresenta la risposta dal master.
type ClientResponse struct {
	Ack        string // Messaggio di conferma
	SortedData []int  // Array ordinato restituito dal master
}

// WorkerRequest rappresenta la richiesta inviata dal master ai worker.
type WorkerRequest struct {
	Chunk     []int // Il chunk di dati da elaborare
	Intervals []int // Gli intervalli su cui basarsi
	Id        int   // ID univoco del worker

	MapperPorts  []string // porte dei worker
	ReducerPorts []string // porte dei reducer

	WorkerRole string // stringa per comunicare al worker il suo ruolo
}

// WorkerResponse rappresenta la risposta del worker al master.
type WorkerResponse struct {
	SortedReducedData []int // I dati ordinati o ridotti restituiti dal worker
}

// WorkerData rappresenta i dati inviati da un mapper ad un reducer tramite RPC
type WorkerData struct {
	SourceWorkerID int   // ID del worker che invia i dati
	TargetRangeID  int   // ID del range a cui appartengono i dati
	Values         []int // Dati da trasferire
}

// MapperPorts gestisce dinamicamente le porte dei worker.
type MapperPorts struct {
	Mu         sync.Mutex // Mutex per la sincronizzazione
	MapperPort []string   // Slice delle porte dei worker
}

type ReducerPorts struct {
	Mu          sync.Mutex // Mutex per la sincronizzazione
	ReducerPort []string   // Slice delle porte dei reducer
}

// GlobalMapperPorts è l'istanza globale di MapperPorts, usata per gestire i worker disponibili.
var GlobalMapperPorts = &MapperPorts{}
var GlobalReducerPorts = &ReducerPorts{}

// AddWorker aggiunge un worker alla lista.
func (wp *MapperPorts) AddWorker(port string) {
	wp.Mu.Lock()
	defer wp.Mu.Unlock()
	wp.MapperPort = append(wp.MapperPort, port)
	//fmt.Println("Worker aggiunto:", port) // Debug
}

func (wp *ReducerPorts) AddReducer(port string) {
	wp.Mu.Lock()
	defer wp.Mu.Unlock()
	wp.ReducerPort = append(wp.ReducerPort, port)
	//fmt.Println("Reducer aggiunto:", port) // Debug
}

// GetPorts restituisce una copia delle porte dei worker attualmente disponibili.
func (wp *MapperPorts) GetPorts() []string {
	wp.Mu.Lock()
	defer wp.Mu.Unlock()
	return append([]string{}, wp.MapperPort...) // Restituisce una copia sicura
}

// GetReducerPorts restituisce una copia delle porte dei reducer attualmente disponibili.
func (wp *ReducerPorts) GetReducerPorts() []string {
	wp.Mu.Lock()
	defer wp.Mu.Unlock()
	return append([]string{}, wp.ReducerPort...) // Restituisce una copia sicura
}
