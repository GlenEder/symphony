module maestro

go 1.26.4

require (
	github.com/gorilla/websocket v1.5.3
	github.com/sstraus/toon_go/toon v1.0.0
)

replace github.com/sstraus/toon_go/toon v1.0.0 => ./lib/toon
