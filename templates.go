package geard

import (
	"text/template"
)

type containerUnit struct {
	Gear            Identifier
	Image           string
	PortSpec        string
	Slice           string
	User            string
	ReqId           string
	GearBasePath    string
	HomeDir         string
	EnvironmentPath string
}

var containerUnitTemplate = template.Must(template.New("unit.service").Parse(`
[Unit]
Description=Gear container {{.Gear}}

[Service]
Type=simple
Slice={{.Slice}}
{{ if .EnvironmentPath }}EnvironmentFile={{.EnvironmentPath}}{{ end }}
ExecStartPre=/bin/bash -c "/usr/bin/getent passwd 'gear-{{.Gear}}'; if [ $? -ne 0 ]; then /usr/sbin/useradd 'gear-{{.Gear}}' -d {{.HomeDir}} -m; fi"
ExecStartPre=/usr/sbin/restorecon -rv {{.HomeDir}}
ExecStartPre={{.GearBasePath}}/bin/geard-util gen-authorized-keys "{{.Gear}}"
ExecStart=/usr/bin/docker run -name "gear-{{.Gear}}" -volumes-from "gear-{{.Gear}}" -a stdout -a stderr {{.PortSpec}} -rm "{{.Image}}"

[Install]
WantedBy=gear.target

# Gear information
X-GearId={{.Gear}}
X-ContainerImage={{.Image}}
X-ContainerUserId={{.User}}
X-ContainerRequestId={{.ReqId}}
`))

type sliceUnit struct {
	Name   string
	Parent string
}

var sliceUnitTemplate = template.Must(template.New("unit.slice").Parse(`
[Unit]
Description=Gear slice {{.Name}}

[Slice]
CPUAccounting=yes
MemoryAccounting=yes
MemoryLimit=512M
{{ if .Parent }}Slice={{.Parent}}{{ end }}

[Install]
WantedBy=gear.target
`))

type targetUnit struct {
	Name     string
	WantedBy string
}

var targetUnitTemplate = template.Must(template.New("unit.target").Parse(`
[Unit]
Description=Gear target {{.Name}}

[Install]
WantedBy={{.WantedBy}}
`))
