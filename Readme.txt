MapReduce

Spiegazione Client

1. Funzionalità:
   - Genera 100 numeri casuali tra 0 e 999.
   - Si connette al server Master tramite RPC sulla porta 1234.
   - Invia i numeri generati al server per l'elaborazione.
   - Riceve e visualizza la risposta del server.

2. Requisiti:
   - Go >= 1.16.
   - Server Master attivo sulla porta 1234.
   - Modulo "MapReduce/utils" per le strutture di richiesta e risposta.

3. Esecuzione:
   - Assicurarsi che il Master e i Worker siano tutti attivi al momento della chiamata, altrimenti questa fallirà.
   - Eseguire il comando: `go run .\client\`.

4. Dettagli Implementativi:
   - I numeri casuali sono generati usando la funzione `rand.Intn`.
   - Il metodo RPC chiamato è "Master.HandleRequest".


Spiegazione Master

1. Funzionalità:
   - Divide i dati in intervalli e chunks.
   - Distribuisce i chunks ai Mapper e Reducer tramite RPC.
   - Coordina l'elaborazione tra i worker Mapper e Reducer.

2. Requisiti:
   - Go >= 1.16.
   - Modulo "MapReduce/utils" per strutture e funzioni ausiliarie.
   - Configurazione dei worker Mapper e Reducer.

3. Configurazione:
   - Mapper: Porte 1230-1233.
   - Reducer: Porte 1235-1239 ESCLUSA la 1238 nel mio caso perchè già in uso.
   - Porta Master: 1234.

4. Esecuzione:
   - Avviare il server con `go run go run .\master\ `.
   - I client devono chiamare "Master.HandleRequest" per inviare dati.

5. Dettagli Implementativi:
   - Utilizza `SampleData` per calcolare gli intervalli.
   - Distribuisce i dati con `DistributeData` su più worker.

Spiegazione Worker

1. Funzionalità:
   - Processa dati ricevuti da un Master (ruolo "Mapper").
   - Suddivide i dati in base a intervalli (Fase di mapping) e li ordina.
   - Trasferisce i dati elaborati ai Worker con il ruolo "Reducer" (fase di Shuffling).
   - I reducer ordinano i dati ricevuti e li scrive su file.

2. Requisiti:
   - Go >= 1.16.
   - Modulo "MapReduce/utils" per supporto funzionale.
   - Master e altri Worker attivi nel sistema.

3. Configurazione:
   - Porta del Worker specificata come argomento alla riga di comando.
     Esempio: ` go run .\worker\ 1235`.
     NOTA: Poichè l'output dei reducer viene riversato su un file, è consigliabile avviare i worker nella root directory
           per evitare problemi con l apertura dei file di output

4. Ruoli Supportati:
   - Mapper: Elabora i chunk ricevuti dal Master, mappandoli e ordinandoli e li distribuisce ai Reducer responsabili per
             una determinata chiave.
   - Reducer: Riceve dati ordinati da Mapper, li ordina e li scrive su file.

5. File di Output:
   - Ogni Worker Reducer scrive i dati elaborati in file come:
     `./Output/Output<N>.txt` (dove <N> è l'ID del Reducer).

6. Esecuzione:
   - Avviare il Worker con `go run .\worker\ 123X', dove X appartiene all'insieme[1230,1231,1232,1233] per i Mapper e
     [1235,1236,1238,1239] per i reducer.
   - Utilizzare il Master per coordinare i dati.

7. Dettagli Implementativi:
   - Suddivide i dati con `mapDataToRanges`.
   - Scrive i risultati con `AppendToFile`.

