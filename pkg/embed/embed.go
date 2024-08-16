package embed

import (
	"embed"

	"github.com/spectrocloud-labs/embeddedfs/pkg/embeddedfs"
)

//go:embed resources/*
var resources embed.FS

// EFS is the nodeadm provider's embedded file system
var EFS = embeddedfs.NewEmbeddedFS("resources", resources)
