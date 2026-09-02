package doelbinding["identificatie-en-ingezetenschap"]

import rego.v1

default allow := false

default reason := ""

consumer_gemeente := data.entities.app[input.subject.id].gemeente

is_brp_request if input.context.http.path == "/v1/haalcentraal/api/brp/personen"

allow if not is_brp_request

is_government_agency if input.subject.attributes.tags[_] == "government-agency"

has_correct_doelbinding if input.context.doelbinding == "identificatie-en-ingezetenschap"

queried_gemeente := data.entities.burger[bsn].gemeente if {
	m := regex.find_all_string_submatch_n(`persoon\.bsn=([0-9]+)`, input.context.body["@filter"], 1)[0]
	bsn := m[1]
}

queried_gemeente := data.entities.adres[postcode].gemeente if {
	m := regex.find_all_string_submatch_n(`adres\.postcode=([0-9A-Z]{6})`, input.context.body["@filter"], 1)[0]
	postcode := m[1]
}

allow if {
	is_brp_request
	is_government_agency
	has_correct_doelbinding
	queried_gemeente == consumer_gemeente
}

reason := "geen overheidsinstantie" if {
	is_brp_request
	not is_government_agency
}

reason := "verkeerde doelbinding" if {
	is_brp_request
	is_government_agency
	not has_correct_doelbinding
}

reason := "inwoner niet van deze gemeente" if {
	is_brp_request
	is_government_agency
	has_correct_doelbinding
	queried_gemeente != consumer_gemeente
}
