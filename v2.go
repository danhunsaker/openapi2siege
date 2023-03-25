package main

import (
	"fmt"

	"github.com/pb33f/libopenapi"
	v2 "github.com/pb33f/libopenapi/datamodel/high/v2"
	"github.com/urfave/cli/v2"
)

func handleV2Spec(c *cli.Context, spec *libopenapi.DocumentModel[v2.Swagger]) (urlList, *SiegeConfig, error) {
	return nil, nil, fmt.Errorf("OpenAPI 2 (Swagger) documents are not yet implemented.\nTHIS WILL HAPPEN IN A FUTURE RELEASE.\n")
}
