package templates

import cterror "github.com/maplain/control-tower/pkg/error"

const (
	TemplateNotSupportedError = cterror.Error("template type not supported")
)

var (
	SupportedTemplateType = map[string]string{
		"build-tile":              BuildTileTemplate,
		"install-tile":            InstallTileTemplate,
		"nsx-acceptance-tests":    NsxAcceptanceTestsTemplate,
		"kubo":                    DeployKuboPipelineTemplate,
		"releng-acceptance-tests": RelengAcceptanceTestsPipelineTemplate,
		"build-pks-nsx-t-release": BuildPksNSXTReleaseTemplate,
	}
)
