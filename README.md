# Papernet

Papernet is a very simple tool to keep track of the papers you read.

## Setup

### Graphviz
On linux
```bash
sudo apt install graphviz
```

On Mac, with homebrew:
```bash
brew install graphviz
```

### Go
Usual go installation (1.6.1)

## Project architecture
```
papernet
|__ data
|   |__ papernet.bolt.db
|   |__ papernet.cayley.db
|__ db
|  |__ paper.go
|__ dot
|   |__ graph.go
|__ models
|   |__ paper.go
|__ public
|   |__ css
|   |   |__ ...
|   |__ fonts
|   |   |__ ...
|   |__ images
|   |   |__ ...
|   |__ templates
|       |__ ...
|__ web
|   |__ router.go
|__ main.go <-- start the server
```
