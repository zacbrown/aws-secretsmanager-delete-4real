package main

import (
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/k0kubun/pp/v3"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

// This PrettyPrinter is used to print out the SDL AST for debugging purposes.
// It removes the colorization so that the output is more easily diffable.
func getColorlessPrettyPrinter() *pp.PrettyPrinter {
	thisPrettyPrinter := pp.New()
	thisPrettyPrinter.SetColoringEnabled(false)
	return thisPrettyPrinter
}

func main() {
	app := &cli.App{
		Name:   "aws-secretsmanager-delete-4real",
		Usage:  "delete an AWS SecretsManager secret 4 real",
		Action: run,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "secret-id",
				Aliases:  []string{"s"},
				Usage:    "Delete secret with identifier `SECRET-ID`",
				Required: true,
			},
			&cli.BoolFlag{
				Name:     "restore-first",
				Usage:    "Restore the secret before deleting it.",
				Aliases:  []string{"r"},
				Required: false,
			},
			&cli.BoolFlag{
				Name:     "verbose",
				Usage:    "Verbose responses from AWS API.",
				Aliases:  []string{"v"},
				Required: false,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func delete(ctx *cli.Context, client *secretsmanager.Client) error {
	secretId := ctx.String("secret-id")
	forceDelete := true

	response, err := client.DeleteSecret(ctx.Context, &secretsmanager.DeleteSecretInput{
		SecretId:                   &secretId,
		ForceDeleteWithoutRecovery: &forceDelete,
	})
	if err != nil {
		return errors.Wrap(err, "failed to delete secret")
	}

	log.Printf("Deletion successful: name='%s', arn='%s', deletion-date='%s'\n",
		*response.Name,
		*response.ARN,
		response.DeletionDate.String(),
	)
	if ctx.Bool("verbose") {
		log.Printf("Deletion raw response: %s\n", getColorlessPrettyPrinter().Sprint(response))
	}

	return nil
}

func restore(ctx *cli.Context, client *secretsmanager.Client) error {
	secretId := ctx.String("secret-id")

	response, err := client.RestoreSecret(ctx.Context, &secretsmanager.RestoreSecretInput{
		SecretId: &secretId,
	})
	if err != nil {
		return errors.Wrap(err, "failed to restore secret")
	}

	log.Printf("Restore successful: name='%s', arn='%s'\n", *response.Name, *response.ARN)
	if ctx.Bool("verbose") {
		log.Printf("Restore raw response: %s\n", getColorlessPrettyPrinter().Sprint(response))
	}

	return nil
}

func run(ctx *cli.Context) error {
	awsConfig, err := config.LoadDefaultConfig(ctx.Context)
	if err != nil {
		return errors.Wrap(err, "failed to load AWS config")
	}
	client := secretsmanager.NewFromConfig(awsConfig)

	if ctx.Bool("restore-first") {
		restore(ctx, client)
	}

	return delete(ctx, client)
}
