package memory

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/store"
)

func loadFDS(s store.Storage) {
	if err := s.AddTableFromCSV("brp", "persoon", personen); err != nil {
		panic(err)
	}
	if err := s.AddTableFromCSV("brp", "adres", adressen); err != nil {
		panic(err)
	}
	if err := s.AddTableFromCSV("rdw", "kenteken", kentekens); err != nil {
		panic(err)
	}
	if err := s.AddTableFromCSV("rdw", "adres", adressen); err != nil {
		panic(err)
	}
}

var personen = [][]string{
	{"bsn", "voornaam", "achternaam"},
	{"999990251", "Pieter Koen", "Wezichem"},
	{"999990263", "Gerrit", "Bommels"},
	{"999990275", "Amaria", "Bont"},
	{"999990287", "Ursula", "Koenders-van Zanten"},
}

var adressen = [][]string{
	{"bsn", "postcode"},
	{"999990251", "1111AA"},
	{"999990263", "2222BB"},
	{"999990275", "3333CC"},
	{"999990287", "4444DD"},
}

var kentekens = [][]string{
	{"kenteken", "bsn"},
	{"AA-11-BB", "999990251"},
	{"CC-22-DD", "999990263"},
	{"EE-22-DD", "999990275"},
	{"BB-11-AA", "999990287"},
}
