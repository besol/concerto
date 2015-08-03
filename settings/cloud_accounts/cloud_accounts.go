package cloud_accounts

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/flexiant/concerto/utils"
	"github.com/flexiant/concerto/webservice"
	"os"
	"text/tabwriter"
)

type Account struct {
	Id          string `json:"id"`
	CloudProvId string `json:"cloud_provider_id"`
}

type RequiredCredentials struct {
	AccessKeyId       string `json:"access_key_id,omitempty"`
	SecretAccessKey   string `json:"secret_access_key,omitempty"`
	Username          string `json:"username,omitempty"`
	APIKey            string `json:"api_key,omitempty"`
	Password          string `json:"password"`
	User              string `json:"user"`
	ClientId          string `json:"client_id,omitempty"`
	GoogleProj        string `json:"google_project,omitempty"`
	GoogleClientEmail string `json:"google_client_email,omitempty"`
	CertGoogleKey     string `json:"cert_google_key,omitempty"`
	SubscriptionID    string `json:"subscription_id,omitempty"`
	CertManagement    string `json:"cert_management_certificate,omitempty"`
}

func cmdList(c *cli.Context) {
	var accounts []Account

	webservice, err := webservice.NewWebService()
	utils.CheckError(err)

	data, err := webservice.Get("/v1/settings/cloud_accounts")
	utils.CheckError(err)

	err = json.Unmarshal(data, &accounts)
	utils.CheckError(err)

	w := tabwriter.NewWriter(os.Stdout, 15, 1, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tCLOUD PROVIDER ID\r")

	for _, ac := range accounts {
		fmt.Fprintf(w, "%s\t%s\n", ac.Id, ac.CloudProvId)
	}

	w.Flush()
}

func cmdCreate(c *cli.Context) {
	utils.FlagsRequired(c, []string{"cloud_provider_id", "credentials"})
	webservice, err := webservice.NewWebService()
	utils.CheckError(err)

	v := make(map[string]interface{})

	v["cloud_provider_id"] = c.String("cloud_provider_id")

	credStr := c.String("credentials")
	recCred := &RequiredCredentials{}
	json.Unmarshal([]byte(credStr), recCred)
	fmt.Println(recCred.Password)

	v["credentials"] = recCred

	jsonBytes, err := json.Marshal(v)
	utils.CheckError(err)
	err, res, _ := webservice.Post("/v1/settings/cloud_accounts", jsonBytes)
	if res == nil {
		log.Fatal(err)
	}
	utils.CheckError(err)

}

func cmdUpdate(c *cli.Context) {
	utils.FlagsRequired(c, []string{"id"})
	webservice, err := webservice.NewWebService()
	utils.CheckError(err)

	v := make(map[string]string)

	if c.IsSet("credentials") {
		v["credentials"] = c.String("credentials")
	}

	jsonBytes, err := json.Marshal(v)
	utils.CheckError(err)
	err, res, _ := webservice.Put(fmt.Sprintf("/v1/settings/cloud_accounts/%s", c.String("id")), jsonBytes)

	utils.CheckError(err)
	fmt.Println(res)
}

func cmdDelete(c *cli.Context) {
	utils.FlagsRequired(c, []string{"id"})

	webservice, err := webservice.NewWebService()
	utils.CheckError(err)

	err, _, res := webservice.Delete(fmt.Sprintf("/v1/settings/cloud_accounts/%s", c.String("id")))
	utils.CheckError(err)
	utils.CheckReturnCode(res)

	fmt.Println(res)
}

func SubCommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "list",
			Usage:  "Lists the cloud accounts of the account group.",
			Action: cmdList,
		},
		{
			Name:   "create",
			Usage:  "Creates a new cloud account.",
			Action: cmdCreate,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "cloud_provider_id",
					Usage: "Identifier of the cloud provider",
				},
				cli.StringFlag{
					Name:  "credentials",
					Usage: "A mapping assigning a value to each of the required yes credentials of the cloud provider (JSON String)",
				},
			},
		},
		{
			Name:   "update",
			Usage:  "Updates an existing cloud account.",
			Action: cmdUpdate,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "id",
					Usage: "Account Id",
				},
				cli.StringFlag{
					Name:  "credentials",
					Usage: "A mapping assigning a value to each of the required yes credentials of the cloud provider (JSON String)",
				},
			},
		},
		{
			Name:   "delete",
			Usage:  "Deletes a cloud account",
			Action: cmdDelete,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "id",
					Usage: "Account Id",
				},
			},
		},
	}
}
