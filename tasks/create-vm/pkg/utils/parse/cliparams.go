package parse

import (
	"github.com/suomiy/kubevirt-tekton-tasks/tasks/create-vm/pkg/constants"
	errors2 "github.com/suomiy/kubevirt-tekton-tasks/tasks/create-vm/pkg/errors"
	"github.com/suomiy/kubevirt-tekton-tasks/tasks/create-vm/pkg/utils"
	"github.com/suomiy/kubevirt-tekton-tasks/tasks/create-vm/pkg/utils/logger"
	"github.com/suomiy/kubevirt-tekton-tasks/tasks/create-vm/pkg/utils/output"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"strings"
)

const (
	vmNamespaceOptionName       = "vm-namespace"
	templateNamespaceOptionName = "template-namespace"
)

const templateParamSep = ":"

// TemplateNamespaces and VirtualMachineNamespaces: arrays allow to have these options without option argument
type CLIOptions struct {
	TemplateName              string            `arg:"--template-name,required" placeholder:"NAME" help:"Name of a template to create VM from"`
	TemplateNamespaces        []string          `arg:"--template-namespace" placeholder:"NAMESPACE" help:"Namespace of a template to create VM from"`
	TemplateParams            []string          `arg:"--template-params" placeholder:"KEY2:VAL1 KEY2:VAL2" help:"Template params to pass when processing the template manifest"`
	VirtualMachineNamespaces  []string          `arg:"--vm-namespace" placeholder:"NAMESPACE" help:"Namespace where to create the VM"`
	DataVolumes               []string          `arg:"--dvs" placeholder:"DV1 DV2" help:"Add DataVolumes to VM Volumes"`
	OwnDataVolumes            []string          `arg:"--own-dvs" placeholder:"DV1 DV2" help:"Add DataVolumes to VM Volumes and add VM to DV ownerReferences. These DVs will be deleted once the created VM gets deleted."`
	PersistentVolumeClaims    []string          `arg:"--pvcs" placeholder:"PVC1 PVC2" help:"Add PersistentVolumeClaims to VM Volumes."`
	OwnPersistentVolumeClaims []string          `arg:"--own-pvcs" placeholder:"PVC1 PVC2" help:"Add PersistentVolumeClaims to VM Volumes and add VM to PVC ownerReferences. These PVCs will be deleted once the created VM gets deleted."`
	Output                    output.OutputType `arg:"-o" placeholder:"FORMAT" help:"Output format. One of: yaml|json"`
	Debug                     bool              `arg:"--debug" help:"Sets DEBUG log level"`
}

func (c *CLIOptions) GetAllPVCNames() []string {
	return utils.ConcatStringSlices(c.OwnPersistentVolumeClaims, c.PersistentVolumeClaims)
}

func (c *CLIOptions) GetAllDVNames() []string {
	return utils.ConcatStringSlices(c.OwnDataVolumes, c.DataVolumes)
}

func (c *CLIOptions) GetAllDiskNames() []string {
	return utils.ConcatStringSlices(c.GetAllPVCNames(), c.GetAllDVNames())
}

func (c *CLIOptions) GetTemplateParams() map[string]string {
	result := make(map[string]string, len(c.TemplateParams))

	for _, keyVal := range c.TemplateParams {
		split := strings.SplitN(keyVal, templateParamSep, 2)
		if len(split) == 2 {
			result[split[0]] = split[1]
		}
	}
	return result
}

func (c *CLIOptions) GetDebugLevel() zapcore.Level {
	if c.Debug {
		return zapcore.DebugLevel
	}
	return zapcore.InfoLevel
}

func (c *CLIOptions) GetTemplateNamespace() string {
	return utils.GetLast(c.TemplateNamespaces)
}

func (c *CLIOptions) GetVirtualMachineNamespace() string {
	return utils.GetLast(c.VirtualMachineNamespaces)
}

func (c *CLIOptions) setTemplateNamespace(namespace string) {
	c.TemplateNamespaces = []string{namespace}
}

func (c *CLIOptions) setVirtualMachineNamespace(namespace string) {
	c.VirtualMachineNamespaces = []string{namespace}
}

func (c *CLIOptions) Init() error {
	defer logger.GetLogger().Debug("parsed arguments", zap.Reflect("cliOptions", c))

	tempNamespace := c.GetTemplateNamespace()
	vmNamespace := c.GetVirtualMachineNamespace()

	if tempNamespace == "" || vmNamespace == "" {
		activeNamespace, err := constants.GetActiveNamespace()
		if err != nil {
			return errors2.NewMissingRequiredError("%v: %v option is empty", err.Error(), c.getMissingNamespaceOptionNames())
		}
		if tempNamespace == "" {
			c.setTemplateNamespace(activeNamespace)
		}
		if vmNamespace == "" {
			c.setVirtualMachineNamespace(activeNamespace)
		}
	}

	if len(c.TemplateNamespaces) > 1 {
		c.setTemplateNamespace(c.GetTemplateNamespace())
	}
	if len(c.VirtualMachineNamespaces) > 1 {
		c.setVirtualMachineNamespace(c.GetVirtualMachineNamespace())
	}

	return nil
}

func (c *CLIOptions) getMissingNamespaceOptionNames() string {
	var result = make([]string, 0, 2)
	if c.GetTemplateNamespace() == "" {
		result = append(result, templateNamespaceOptionName)
	}
	if c.GetVirtualMachineNamespace() == "" {
		result = append(result, vmNamespaceOptionName)
	}

	return strings.Join(result, "/")
}
