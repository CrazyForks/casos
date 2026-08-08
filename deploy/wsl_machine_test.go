package deploy

import "testing"

func TestLocalWSLMachineName(t *testing.T) {
	cases := map[string]string{
		"Ubuntu":            "wsl-ubuntu",
		"Ubuntu-22.04":      "wsl-ubuntu-22-04",
		"openSUSE Leap 15":  "wsl-opensuse-leap-15",
		"  Debian  ":        "wsl-debian",
		"":                  "wsl-default",
		"___":               "wsl-default",
		"Ubuntu_20.04_LTS_": "wsl-ubuntu-20-04-lts",
	}
	for distro, want := range cases {
		if got := localWSLMachineName(distro); got != want {
			t.Errorf("localWSLMachineName(%q) = %q, want %q", distro, got, want)
		}
	}
}

func TestLocalWSLDisplayName(t *testing.T) {
	if got := localWSLDisplayName(" Ubuntu "); got != "WSL: Ubuntu" {
		t.Errorf("localWSLDisplayName = %q", got)
	}
	if got := localWSLDisplayName(""); got != "Local WSL" {
		t.Errorf("localWSLDisplayName = %q", got)
	}
}
