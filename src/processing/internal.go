package processing

import (
	"fmt"
	"os/exec"
	"strings"
)

func WhoisParameters(ip string) (string, string) {
	app := "whois"
	cmd := fmt.Sprintf("%s %s |grep country -i -m 1 |cut -d ':' -f 2 |xargs", app, ip)
	country, err := exec.Command("bash", "-c", cmd).Output()
	cmd2 := fmt.Sprintf("%s %s |grep organization -i -m 1 |cut -d ':' -f 2 |xargs", app, ip)
	organization, err := exec.Command("bash", "-c", cmd2).Output()

	if err != nil {
		return "", ""
	}

	return strings.TrimSuffix(string(country), "\n"), strings.TrimSuffix(string(organization), "\n")
}
