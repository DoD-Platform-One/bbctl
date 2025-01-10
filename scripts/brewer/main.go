package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"text/template"

	"repo1.dso.mil/big-bang/product/packages/bbctl/scripts/brewer/fetcher"
)

const bbctlProjectID = 11320
const templatesDir = "./scripts/brewer/templates"
const templateName = "bbctl.rb.tmpl"

func main() {
	rf := fetcher.NewReleaseFetcher(bbctlProjectID)
	err := generateBbctlTemplate(rf)
	if err != nil {
		log.Fatal(fmt.Errorf("brewer failed to generate a homebrew formula: %w", err))
	}
}

func renderBrewFormula(w io.Writer, releaseInfo fetcher.ReleaseInfo, templateName string) error {
	templatePath := path.Join(templatesDir, templateName)
	absPath, err := filepath.Abs(templatePath)
	if err != nil {
		return err
	}
	tmpl, err := template.ParseFiles(absPath)
	if err != nil {
		return err
	}

	err = tmpl.Execute(w, releaseInfo)
	if err != nil {
		return err
	}

	return nil
}

func generateBbctlTemplate(glrf *fetcher.ReleaseFetcher) error {
	release, err := glrf.FetchLatestReleaseInfo()
	if err != nil {
		return fmt.Errorf("unable to fetch latest release info: %w", err)
	}

	sourceTarballURI, err2 := release.SourceTarballURI()
	if err2 != nil {
		return fmt.Errorf("unable to extract source tarball URI: %w", err2)
	}

	tarballBytes, err := glrf.FetchRepo1Uri(sourceTarballURI)
	if err != nil {
		return fmt.Errorf("unable to fetch repo1 URI: %w", err)
	}
	hash := sha256.New()
	hash.Write(tarballBytes)
	sha256Sum := hex.EncodeToString(hash.Sum(nil))

	releaseInfo := fetcher.ReleaseInfo{
		ReleaseTag: release.TagName,
		ReleaseURI: sourceTarballURI,
		Sha256Sum:  sha256Sum,
	}

	err4 := renderBrewFormula(os.Stdout, releaseInfo, templateName)
	if err4 != nil {
		return fmt.Errorf("unable to render release info template: %w", err4)
	}

	return nil
}
