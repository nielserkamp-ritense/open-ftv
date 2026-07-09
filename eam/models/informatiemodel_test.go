package models

import "testing"

func TestNormalizeActionURI(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty", in: "", want: ""},
		{name: "avg raadplegen", in: "raadplegen", want: ActieRaadplegen},
		{name: "avg with spaces", in: "beschikbaar stellen", want: ActieBeschikbaarStellen},
		{name: "multi word", in: "met elkaar in verband brengen", want: ActieInVerbandBrengen},
		{name: "synonym registreren", in: "registreren", want: ActieVastleggen},
		{name: "synonym verwijderen", in: "verwijderen", want: ActieWissen},
		{name: "synonym doorleveren", in: "doorleveren", want: ActieDoorzenden},
		{name: "case insensitive", in: "  RaadPlegen ", want: ActieRaadplegen},
		{name: "http method get", in: "GET", want: ActieRaadplegen},
		{name: "http method delete", in: "DELETE", want: ActieWissen},
		{name: "already uri", in: ActieRaadplegen, want: ActieRaadplegen},
		{name: "unknown unchanged", in: "can_read", want: "can_read"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := NormalizeActionURI(tc.in); got != tc.want {
				t.Fatalf("NormalizeActionURI(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestNormalizeSubjectTypeURI(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   string
		want string
	}{
		{name: "user", in: "user", want: SubjectUser},
		{name: "app", in: "app", want: SubjectApp},
		{name: "doelbinding", in: "doelbinding", want: SubjectDoelbinding},
		{name: "activity", in: "activity", want: SubjectActivity},
		{name: "zaak", in: "zaak", want: SubjectZaak},
		{name: "fsc-peer", in: "fsc-peer", want: SubjectFSCPeer},
		{name: "ip-address", in: "ip-address", want: SubjectIPAddress},
		{name: "ip_address alias", in: "ip_address", want: SubjectIPAddress},
		{name: "already uri", in: SubjectUser, want: SubjectUser},
		{name: "unknown unchanged", in: "robot", want: "robot"},
		{name: "empty", in: "", want: ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := NormalizeSubjectTypeURI(tc.in); got != tc.want {
				t.Fatalf("NormalizeSubjectTypeURI(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestNormalizeResourceTypeURI(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   string
		want string
	}{
		{name: "resource", in: "resource", want: ResourceResource},
		{name: "service", in: "service", want: ResourceService},
		{name: "uri", in: "uri", want: ResourceURI},
		{name: "zaak", in: "zaak", want: ResourceZaak},
		{name: "already uri", in: ResourceService, want: ResourceService},
		{name: "unknown unchanged", in: "brp-personen", want: "brp-personen"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := NormalizeResourceTypeURI(tc.in); got != tc.want {
				t.Fatalf("NormalizeResourceTypeURI(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestNormalizeEntityTypeURI(t *testing.T) {
	t.Parallel()

	// subject wins over resource for the ambiguous "zaak".
	if got := NormalizeEntityTypeURI("zaak"); got != SubjectZaak {
		t.Fatalf("NormalizeEntityTypeURI(zaak) = %q, want %q", got, SubjectZaak)
	}
	if got := NormalizeEntityTypeURI("service"); got != ResourceService {
		t.Fatalf("NormalizeEntityTypeURI(service) = %q, want %q", got, ResourceService)
	}
	if got := NormalizeEntityTypeURI("unknown"); got != "unknown" {
		t.Fatalf("NormalizeEntityTypeURI(unknown) = %q, want unchanged", got)
	}
}
