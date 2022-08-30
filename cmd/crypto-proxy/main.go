package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	cryptoproxy "homework.com/crypro-proxy"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	if err := cryptoproxy.CreateCommand().Execute(); err != nil {
		// TODO: add error codes to improve readability in cases of errors
		log.Fatal().Err(err).Msg("")
	}
}
