package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"sync"
)

type Record struct {
	Value  []byte `json:"value"`
	Offset uint64 `json:"offset"`
}

type Log struct {
	records []Record
	mu      sync.Mutex
}

// Estp agrega un nuevo registro al log y devuelve el offset del registro que se agrego
func (l *Log) AddRecord(record Record) uint64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	record.Offset = uint64(len(l.records))
	l.records = append(l.records, record)
	return record.Offset
}

// Obtiene un registro del log usando el offset dado
func (l *Log) GetRecord(offset uint64) (Record, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if offset >= uint64(len(l.records)) {
		return Record{}, false
	}
	return l.records[offset], true
}

// Instancia global de Log para manejar los registros
var log = Log{}

// Maneja las solicitudes a la raíz (o sea '/') y decide si se debe leer o escribir en el log (para q quede como en el ejemplo que no pone /write o /read)
func rootHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		// Para solicitudes POST, leemos el cuerpo de la solicitud y tratamos de desmarshallearlo en un record
		var record Record
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error leyendo el cuerpo: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		if err := json.Unmarshal(body, &record); err != nil {
			http.Error(w, "Error al intentar hacer unmarshaling del JSON: "+err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json") // Intente poner esto basicamente en cualquier lugar del codigo que tuviera sentido pero no se xq no logro que funcione usar el curl sin especificar el content type como en el ejemplo :(

		offset := log.AddRecord(record)
		json.NewEncoder(w).Encode(map[string]uint64{"offset": offset})

	case http.MethodGet:
		// Para solicitudes GET, obtenemos el parámetro offset de la URL para hacer match con los q tenemos
		offsetStr := r.URL.Query().Get("offset")
		if offsetStr == "" {
			http.Error(w, "Se necesita ingresar un Offset", http.StatusBadRequest)
			return
		}

		// Convertimos el offset a un número entero sin signo.
		offset, err := strconv.ParseUint(offsetStr, 10, 64)
		if err != nil {
			http.Error(w, "Offset ingresado es invalido: "+err.Error(), http.StatusBadRequest)
			return
		}

		record, ok := log.GetRecord(offset)
		if !ok {
			http.Error(w, "No se encontro un record con ese offset", http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(record)

	default:
		http.Error(w, "Ese metodo no esta soportado, usa POST o GET, en el README hay ejemplos de q comandos funcionan", http.StatusMethodNotAllowed)
	}
}

func main() {
	http.HandleFunc("/", rootHandler)
	http.ListenAndServe(":8080", nil)
}
