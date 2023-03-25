package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/utils"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
)

func main() {
	app := cli.NewApp()
	app.Name = "openapi2siege"
	app.Version = "0.1.0"
	app.Usage = "OpenAPI to URLs file converter"
	app.Description = "Parses an OpenAPI specification and converts it to a Siege URLs file"
	app.EnableBashCompletion = true
	app.Suggest = true
	app.UseShortOptionHandling = true
	app.HideHelpCommand = true

	app.Authors = []*cli.Author{
		{Name: "Hennik Hunsaker", Email: "hennikhunsaker@gmail.com"},
	}

	app.Flags = []cli.Flag{
		// Either pass all the flags below, or use this handy config file instead!
		&cli.PathFlag{
			Name:      "conf",
			Aliases:   []string{"c"},
			Usage:     "specify the configuration `file` to use",
			Value:     "oa2s.conf",
			TakesFile: true,
		},
		// OpenAPI-related configs
		altsrc.NewPathFlag(&cli.PathFlag{
			Name:      "spec",
			Aliases:   []string{"s"},
			Usage:     "specify the (root) OpenAPI `file` to convert",
			Value:     "openapi.yaml",
			TakesFile: true,
		}),
		altsrc.NewBoolFlag(&cli.BoolFlag{
			Name:   "server.useFirst",
			Usage:  "use the first server in the OpenAPI spec",
			Hidden: true,
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:   "server.description",
			Usage:  "use the server matching this description in the OpenAPI spec",
			Hidden: true,
		}),
		altsrc.NewGenericFlag(&cli.GenericFlag{
			Name:   "server.variables",
			Usage:  "use the provided variable values when generating the server's baseUrl",
			Value:  ServerVarsConfig{},
			Hidden: true,
		}),
		altsrc.NewGenericFlag(&cli.GenericFlag{
			Name:   "auth",
			Usage:  "configure authentication details for the auth scheme you wish to use",
			Value:  AuthConfig{},
			Hidden: true,
		}),
		altsrc.NewGenericFlag(&cli.GenericFlag{
			Name:   "paths",
			Usage:  "configure path details: paths.{path}.{method}.params.{name}, paths.{path}.post.payloads.{mediatype}",
			Value:  PathsConfig{},
			Hidden: true,
		}),
		// Siege-related configs
		altsrc.NewPathFlag(&cli.PathFlag{
			Name:      "siege.urls",
			Usage:     "specify the `path` of the urls.txt to generate",
			Value:     "urls.txt",
			TakesFile: true,
			Hidden:    true,
		}),
		altsrc.NewPathFlag(&cli.PathFlag{
			Name:      "siege.cookies",
			Usage:     "specify the `path` of the cookies.txt to generate",
			Value:     "cookies.txt",
			TakesFile: true,
			Hidden:    true,
		}),
		altsrc.NewPathFlag(&cli.PathFlag{
			Name:      "siege.config",
			Usage:     "specify the `path` of the siege.conf to generate",
			Value:     "siege.conf",
			TakesFile: true,
			Hidden:    true,
		}),
	}

	app.Before = func(ctx *cli.Context) error {
		if !ctx.IsSet("conf") {
			if err := ctx.Set("conf", "oa2s.conf"); err != nil {
				return err
			}
		}

		return altsrc.InitInputSourceWithContext(app.Flags, altsrc.DetectNewSourceFromFlagFunc("conf"))(ctx)
	}

	app.Action = func(c *cli.Context) error {
		specPath := c.Path("spec")

		specBytes, err := os.ReadFile(specPath)
		if err != nil {
			return fmt.Errorf("Could not read %s\n%v\n", specPath, err)
		}

		specDoc, err := libopenapi.NewDocument(specBytes)
		if err != nil {
			return fmt.Errorf("Could not parse %s\n%v\n", specPath, err)
		}

		var urls urlList
		var conf *SiegeConfig

		switch specDoc.GetSpecInfo().SpecType {
		case utils.OpenApi2:
			specV2, errs := specDoc.BuildV2Model()
			if len(errs) > 0 {
				for _, err = range errs {
					if err != nil {
						fmt.Printf("Could not load v2 spec in %s\n%v\n", specPath, err)
					}
				}

				return fmt.Errorf("Aborting.\n")
			}

			urls, conf, err = handleV2Spec(c, specV2)
			if err != nil {
				return err
			}
		case utils.OpenApi3:
			specV3, errs := specDoc.BuildV3Model()
			if len(errs) > 0 {
				for _, err = range errs {
					if err != nil {
						fmt.Printf("Could not load v3 spec in %s\n%v\n", specPath, err)
					}
				}

				return fmt.Errorf("Aborting.\n")
			}

			urls, conf, err = handleV3Spec(c, specV3)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("Could not load spec in %s\nUnknown Spec Type %s\n", specPath, specDoc.GetSpecInfo().SpecType)
		}

		conf.GetMethod = "GET"

		urlFile := c.Path("siege.urls")
		configFile := c.Path("siege.config")
		cookieFile := c.Path("siege.cookies")

		cookies, err := urls.CookieJar(cookieFile)
		if err != nil {
			return err
		}
		if err = cookies.Save(); err != nil {
			return err
		}

		multipleTypes := false
		if len(urls.MediaTypes()) > 1 {
			multipleTypes = true
		}

		for _, mediaType := range urls.MediaTypes() {
			myUrlFile := urlFile
			myConfigFile := configFile

			if multipleTypes {
				splitType := strings.Split(mediaType, "/")
				furtherSplitType := strings.Split(splitType[len(splitType)-1], "+")
				prefix := furtherSplitType[len(furtherSplitType)-1]

				myUrlFile = prefixFilename(prefix, myUrlFile)
				myConfigFile = prefixFilename(prefix, myConfigFile)
			}

			if err = os.WriteFile(myUrlFile, []byte(urls.StringByMediaType(mediaType)), os.ModePerm); err != nil {
				return err
			}

			conf.UrlFile = myUrlFile

			if err = os.WriteFile(myConfigFile, []byte(conf.String()), os.ModePerm); err != nil {
				return err
			}

			fmt.Printf("\nConversion complete! To use, run\n\tsiege -R %s -T '%s'\n", myConfigFile, mediaType)
		}

		fmt.Println("")

		return nil
	}

	err := app.Run(os.Args)

	if err != nil {
		log.Fatal(err)
	}
}

func prefixFilename(prefix, filename string) string {
	dir, file := path.Split(filename)

	prefixedFile := fmt.Sprintf("%s.%s", prefix, file)

	return path.Join(dir, prefixedFile)
}
