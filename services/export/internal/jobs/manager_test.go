package jobs

import "testing"

func TestTransicaoValida(t *testing.T) {
	validas := [][2]string{
		{StatusCriado, StatusColetando},
		{StatusCriado, StatusErro},
		{StatusColetando, StatusPronto},
		{StatusColetando, StatusErro},
		{StatusPronto, StatusExpirado},
		{StatusCriado, StatusExpirado},
	}
	for _, tr := range validas {
		if !TransicaoValida(tr[0], tr[1]) {
			t.Errorf("%s → %s deveria ser válida", tr[0], tr[1])
		}
	}

	invalidas := [][2]string{
		{StatusCriado, StatusPronto},   // não pode pular coleta
		{StatusPronto, StatusColetando}, // terminal (exceto expirar)
		{StatusErro, StatusColetando},   // terminal
		{StatusExpirado, StatusPronto},  // terminal
		{StatusColetando, StatusCriado}, // não regride
	}
	for _, tr := range invalidas {
		if TransicaoValida(tr[0], tr[1]) {
			t.Errorf("%s → %s deveria ser inválida", tr[0], tr[1])
		}
	}
}
