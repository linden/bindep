package bindep

import "testing"

func TestBinDep(t *testing.T) {
	dep, err := New("https://github.com/linden/btcd.git","36d61891fdd6781a6d0d306cf7d38032afc333f9", nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%s\n", dep)
}
