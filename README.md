
# MapReduce Distributed System

Questo progetto implementa un sistema distribuito basato sul modello **MapReduce**. Utilizza il linguaggio **Go** e comunica tramite **RPC** per coordinare il lavoro tra un **Client**, un **Master** e diversi **Worker**. Ogni componente ha ruoli e responsabilità specifici, e i dati vengono suddivisi in chunk, elaborati e poi ridotti in un file di output.

## Indice dei Contenuti

- [Overview](#overview)
- [Client](#client)
    - [Funzionalità](#client-funzionalità)
    - [Requisiti](#client-requisiti)
    - [Esecuzione](#client-esecuzione)
    - [Dettagli Implementativi](#client-dettagli-implementativi)
- [Master](#master)
    - [Funzionalità](#master-funzionalità)
    - [Requisiti](#master-requisiti)
    - [Configurazione](#master-configurazione)
    - [Esecuzione](#master-esecuzione)
    - [Dettagli Implementativi](#master-dettagli-implementativi)
- [Worker](#worker)
    - [Funzionalità](#worker-funzionalità)
    - [Requisiti](#worker-requisiti)
    - [Configurazione](#worker-configurazione)
    - [Ruoli Supportati](#worker-ruoli-supportati)
    - [File di Output](#worker-file-di-output)
    - [Esecuzione](#worker-esecuzione)
    - [Dettagli Implementativi](#worker-dettagli-implementativi)
- [Workflow](#workflow)
    - [Dettaglio delle Chiamate](#workflow-dettaglio-delle-chiamate)

## Overview

Il progetto implementa un sistema distribuito che simula il funzionamento di **MapReduce**. Il flusso di lavoro è il seguente:

1. **Client** genera un vettore di 100 interi casuali e li invia al **Master**.
2. Il **Master** coordina l'elaborazione tra diversi **Worker**, suddividendo il lavoro tra **Mapper** e **Reducer**.
3. I **Mapper** ricevono i dati, effettuano il mapping e li inviano ai **Reducer**.
4. I **Reducer** aggregano, ordinano e salvano i risultati in file di output.

## Client

### Funzionalità

- Genera 100 numeri casuali compresi tra 0 e 999.
- Si connette al **Master** in ascolto tramite RPC sulla porta **1234**.
- Invia i numeri generati al **Master** per l'elaborazione.

### Requisiti

- Linguaggio **Go >= 1.16**.
- Server **Master** attivo sulla porta **1234**.
- Modulo **MapReduce/utils** per la gestione delle richieste RPC.

### Esecuzione

1. Avviare il **Master** e i **Worker**.
2. Esegui il client con il seguente comando:

```bash
go run ./client/
```

### Dettagli Implementativi

- Generazione dei numeri casuali tramite `rand.Intn`.
- Utilizzo della chiamata RPC `Master.HandleRequest` per comunicare con il **Master**.

## Master

### Funzionalità

- Precampiona una piccola parte dei dati ricevuti dal **Client** per determinare dei range di valori
- Divide i dati ricevuti dal **Client** in chunk.
- Assegna i chunk ai **Mapper**.
- Coordina la fase di mapping e shuffling tra **Mapper** e **Reducer**.

### Requisiti

- Linguaggio **Go >= 1.16**.
- Modulo **MapReduce/utils** per strutture e funzioni ausiliarie.

### Configurazione

- Mapper: Porte **1230-1233**.
- Reducer: Porte **1235-1239** (escluso la porta **1238** poichè nel mio caso già in uso).
- Porta **Master**: **1234**.

### Esecuzione

1. Avvia il **Master** con:

```bash
go run ./master/
```

### Dettagli Implementativi

- Intervalli di dati calcolati con `SampleData`.
- Distribuzione dei task di **Mapper** e **Reducer** gestita tramite `DistributeData`.

## Worker

### Funzionalità

- **Mapper**:
    - Suddivide i dati ricevuti in base agli intervalli.
    - Ordina i dati e li invia ai **Reducer** responsabili per un particolare range di valori.
- **Reducer**:
    - Riceve i dati, li ordina e li salva in un file di output.

### Requisiti

- Linguaggio **Go >= 1.16**.
- Modulo **MapReduce/utils** per supporto funzionale.
- **Master** e altri **Worker** attivi nel sistema.

### Configurazione

- Porta del **Worker** specificata come argomento:
    - Mapper: **1230-1233**.
    - Reducer: **1235-1239** (escluso **1238**).

### Esecuzione

1. Esegui il **Mapper**:

```bash
go run ./worker/ 1230
```

2. Esegui il **Reducer**:

```bash
go run ./worker/ 1235
```

> **Nota**: Avviare i **Worker** dalla root directory per evitare problemi con i file di output.

### Ruoli Supportati

- **Mapper**:
    - Suddivide i chunk in intervalli.
    - Ordina i dati e li invia ai **Reducer**.
- **Reducer**:
    - Riceve i dati dai **Mapper**, li ordina e salva i risultati in file di output.

### File di Output

Ogni **Reducer** salva i risultati in un file denominato:

```text
./Output/Output<N>.txt
```

dove `<N>` è l'ID del **Reducer**.

### Dettagli Implementativi

- Mapping dei dati negli intervalli corretti tramite `mapDataToRanges`.
- Scrittura dei risultati su file tramite `AppendToFile`.

## Workflow

Il flusso di lavoro del sistema segue le seguenti fasi:

1. **Client**:
    - Genera 100 numeri casuali.
    - Chiama il metodo RPC `Master.HandleRequest` per inviare i dati al **Master**.
2. **Master**:
    - Riceve i dati dal **Client**, effettua un piccolo pre-campionamento per determinare degli intervalli usati dai Reducer tramite `SampleData`
    - e li divide in chunk tramite `DistributeData`.
    - Assegna i chunk ai **Mapper** (porte **1230-1233**).
    - Da inizio alla fase di **mapping** e **shuffling** tra **Mapper** e **Reducer chiamando `Worker.ProcessChunk`**.

3. **Mapper**:
    - Suddivide i dati con `mapDataToRanges`.
    - Invia i dati ai **Reducer** tramite RPC (`Worker.ReceiveData`).
4. **Reducer**:
    - Riceve i dati, li aggrega e li ordina.
    - Scrive i risultati in un file di output.

### Dettaglio delle Chiamate RPC

| Componente | Metodo RPC                | Descrizione                                          |
|------------|---------------------------|------------------------------------------------------|
| **Client** | `Master.HandleRequest`     | Invia i dati generati al **Master**.                 |
| **Master** | `Worker.ProcessChunk`      | Distribuisce i chunk ai **Mapper** per la fase di mapping. |
| **Mapper** | `Worker.ReceiveData`       | Invia i dati elaborati ai **Reducer**.               |



