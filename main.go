package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"cloud.google.com/go/storage"
	"github.com/davecgh/go-spew/spew"
	"github.com/urfave/cli"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

var ProjectID string = "core-infra"
var Credentials string = "creds.json"

type Transaction struct {
	Ctx         context.Context
	Client      *storage.Client
	ProjectID   string
	Credentials string
}

func (t *Transaction) New() (*Transaction, error) {
	t.Ctx = context.Background()
	t.ProjectID = "core-infra"
	t.Credentials = "creds.json"

	client, err := storage.NewClient(t.Ctx, option.WithCredentialsFile(t.Credentials))
	if err != nil {
		return nil, err
	}

	t.Client = client

	return t, nil
}

func main() {
	t := &Transaction{}
	t, err := t.New()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	bucketVar := "GOOGLE_CLOUD_STORAGE_BUCKET"
	if os.Getenv(bucketVar) == "" {
		fmt.Printf("Please set %q in your environment to define a bucket to work with.\n", bucketVar)
		os.Exit(2)
	}

	bucket := os.Getenv(bucketVar)

	app := cli.NewApp()
	app.Name = "gstasher"
	app.Usage = "Google Cloud Storage File Manager"
	app.Version = "0.0.1"
	app.Action = func(c *cli.Context) error {
		cli.ShowAppHelp(c)
		return cli.NewExitError("", 2)
	}

	//ctx := context.Background()
	//client, err := storage.NewClient(ctx, option.WithCredentialsFile(Credentials))
	//if err != nil {
	//	log.Fatal(err)
	//}

	app.Commands = []cli.Command{
		{
			Name:    "list",
			Aliases: []string{"l", "ls"},
			Action: func(c *cli.Context) error {
				listBucket(t.Ctx, t.Client, bucket)
				//listBucket(ctx, client, bucket)
				return nil
			},
			Usage: "list files in bucket",
		},
		{
			Name:    "upload",
			Aliases: []string{"u", "up"},
			Action: func(c *cli.Context) error {
				if len(c.Args()) == 0 {
					cli.ShowCommandHelpAndExit(c, "upload", 2)
				}
				uploadFiles(t.Ctx, t.Client, c.Args(), bucket)
				//uploadFiles(ctx, client, c.Args(), bucket)
				return nil
			},
			Usage: "upload files to bucket",
		},
		//{
		//	Name:    "delete",
		//	Aliases: []string{"d", "del"},
		//	Action: func(c *cli.Context) error {
		//		if len(c.Args()) == 0 {
		//			cli.ShowCommandHelpAndExit(c, "delete", 2)
		//		}
		//		deleteFiles(ctx, client, c.Args(), bucket)
		//		return nil
		//	},
		//	Usage: "list files in bucket",
		//},
		//{
		//	Name:    "stat",
		//	Aliases: []string{"st"},
		//	Action: func(c *cli.Context) error {
		//		args := c.Args()
		//		if len(args) == 0 {
		//			cli.ShowCommandHelpAndExit(c, "delete", 2)
		//		}
		//		deleteFiles(ctx, client, c.Args(), bucket)
		//		return nil
		//	},
		//	Usage: "list files in bucket",
		//},
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

	if bucket == "" {
		fmt.Println("must specify a bucket name")
		os.Exit(1)
	}

	//if !isBucketExist(ctx, client, bucket) {
	//	fmt.Printf("bucket %q does not exist.\n", bucket)
	//	os.Exit(1)
	//}

	//listBuckets(ctx, client)
	//createFile(ctx, client, "test.txt", "tourtique")

	//listBucket(ctx, client, bucket)

	//statFile(ctx, client, "test.txt", "tourtique")

}

func isBucketExist(ctx context.Context, client *storage.Client, bucket string) bool {
	it := client.Buckets(ctx, ProjectID)

	for {
		b, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			continue
		}
		if b.Name == bucket {
			return true
		}
	}

	return false
}

func createFile(ctx context.Context, client *storage.Client, fileName, bucket string) {
	b := client.Bucket(bucket)
	wc := b.Object(fileName).NewWriter(ctx)

	if _, err := wc.Write([]byte("12345\n")); err != nil {
		log.Println("unable to write %q to bucket %q", fileName, bucket)
		return
	}

	if err := wc.Close(); err != nil {
		log.Println("unable to close bucket %q", bucket)
		return
	}

}

func uploadFiles(ctx context.Context, client *storage.Client, files []string, bucket string) {
	b := client.Bucket(bucket)

	for _, file := range files {
		f, err := ioutil.ReadFile(file)
		if err != nil {
			fmt.Printf("** Could not open %q: %v\n", file, err)
			continue
		}

		wc := b.Object(file).NewWriter(ctx)

		if _, err := wc.Write(f); err != nil {
			fmt.Printf("** unable to write %q: %v\n", file, err)
			continue
		}

		if err := wc.Close(); err != nil {
			fmt.Printf("** unable to write %q: %v\n", file, err)
			continue
		}
	}
}

func deleteFiles(ctx context.Context, client *storage.Client, files []string, bucket string) {
	b := client.Bucket(bucket)

	for _, file := range files {
		err := b.Object(file).Delete(ctx)
		if err != nil {
			fmt.Printf("** unable to delete %q: %v\n", file, err)
		}
	}

}

func listBucket(ctx context.Context, client *storage.Client, bucket string) {
	it := client.Bucket(bucket).Objects(ctx, nil)
	for {
		obj, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(obj.Name)
	}

}
func listBuckets(ctx context.Context, client *storage.Client) {
	it := client.Buckets(ctx, ProjectID)
	for {
		battrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(battrs.Name)
	}
}

func statFile(ctx context.Context, client *storage.Client, fileName string, bucket string) {
	obj, err := client.Bucket(bucket).Object(fileName).Attrs(ctx)
	if err != nil {
		log.Println(err)
		return
	}

	spew.Dump(obj)

}
