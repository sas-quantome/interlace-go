/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package contractor

import (
	"fmt"
	"path"
	"strings"
	"unicode"

	"github.com/sas-quantome/interlace-go/pkg/cowriter"
	"github.com/sas-quantome/interlace-go/pkg/runner"
	"github.com/sas-quantome/interlace-go/pkg/storekeeper"
)

type Refiner struct {
	Cat      *storekeeper.Cataloguer   // Cat hashes text data and keeps only unique text strings.
	CW       *cowriter.Cowriter        // CW writes and manages intermediate files.
	Base     *runner.Dispenser         // Base dispenses the workspace directory structure and files.
	Mgr      *runner.Manager           // Mgr commits the execution flow and handles resource management.
	Enum     storekeeper.Enumerator    // Enum creates and dynamically updates controlled vocabulary enumerators.
	EventNS  runner.DispenserNamespace // EventNS is the namespace for the event's workflow.
	EntityNS runner.DispenserNamespace // EntityNS is the namespace for the entity's asset.
	binNames runner.Stack[string]      // binNames stores the names of the encoding bin files.
	srcNames []string                  // srcNames is the exact name of the file to download without the URL path.
	urlPaths []string                  // urlPaths stores the full paths of the selected source file names.
	count    int                       // count keeps track of the number of asset records aggregated.
}

// NewRefiner creates a new Refiner instance and returns true to commit work, else the help page is printed.
func NewRefiner(hasURI, askForAsset bool, concept, workflowID, appMode, appName, appAim, appVer string, srcNames []string) (*Refiner, bool) {
	refiner := new(Refiner)
	refiner.Mgr = runner.NewManager()
	// Capture the environment variables e.g., ENV_AEE_BASE_PATH, ENV_ALT_TEMP_PATH, and ENV_ALT_APPS_PATH.
	refiner.Mgr.SetConfigTopic(appMode, appName, appAim, appVer)
	// Declare the exact URL file names to download which must match the filepath base name in the URL input file.
	refiner.srcNames = srcNames
	// Capture the command line arguments.
	refiner.Mgr.ConfigureWork(concept, workflowID, hasURI, askForAsset)
	// Print the help page if requested and return false to exit the program.
	if refiner.Mgr.GetBitFlag(runner.MgrFlgHelp) == true {
		refiner.Mgr.HelpPage(runner.HelpDoc(refiner.srcNames, ""))
		return refiner, false
	}
	// Print the starting information and return true to commit work.
	refiner.Mgr.PrintStartingInfo()
	return refiner, true
}

// Commit returns true if the asset is ready to be committed and any config errors encountered.
func (refiner *Refiner) Commit(assetName string) (bool, error) {
	var err error
	refiner.EventNS, refiner.EntityNS,
		refiner.Base, refiner.urlPaths, err = refiner.Mgr.SetResources(assetName, refiner.srcNames)
	if err != nil {
		return false, err
	}
	if refiner.Mgr.GetBitFlag(runner.MgrFlgNoCommit) == true {
		return false, nil
	}
	return true, nil
}

// ReportCounts uses loggers to report the total number of rows read, records selected, and saved.
func (refiner *Refiner) ReportCounts(lines, processed int, name string) {
	refiner.Mgr.Say(fmt.Sprintf("total %s lines read: %d", name, lines))
	refiner.Mgr.Say(fmt.Sprintf("total %s table rows examined: %d", name, processed))
	refiner.Mgr.Say(fmt.Sprintf("total %s tuples created: %d", name, refiner.count))
}

// Writer formats and writes a string to an internal buffer while incrementing the record counter.
func (refiner *Refiner) Writer(format string, a ...any) error {
	_, err := fmt.Fprintf(&refiner.CW.Buf, format, a...)
	if err != nil {
		return err
	}
	refiner.count++
	return refiner.FlushSeg()
}

// FlushSeg writes the internal buffer to disk if it exceeds 8K.
func (refiner *Refiner) FlushSeg() error {
	return refiner.CW.BytesToDisk(8 * 1024)
}

// FlushAll writes the internal buffer to disk.
func (refiner *Refiner) FlushAll() error {
	return refiner.CW.BytesToDisk(1)
}

// AddBin adds the given bin file name to the binNames stack, it must be in order of the columns in the CW tsv file.
// To skip a CW tsv column, enter any string not starting with the index prefix such as suk_ and udk_
func (refiner *Refiner) AddBin(name string) {
	refiner.binNames.Push(name)
}

// ClearBin removes all bin names from the refiner's internal collection.
func (refiner *Refiner) ClearBin() {
	refiner.binNames.Clear()
}

// InterlaceProperties uses the interlacer to create a bin file from the CS tsv file.
// Assumes the bin file names are in order of the columns in the CS tsv file.
// Invalid names are skipped and considered unwanted column items.
func (refiner *Refiner) InterlaceProperties() error {
	var err error
	interlacer := storekeeper.NewInterlacer(
		8*1024,
		0, // because the asset input file in here is the curated creation of this application.
		refiner.CW.File.Name(),
		refiner.Base.EntityPropertyPath(refiner.EntityNS))
	xType := storekeeper.UDK
	dType := storekeeper.SDT
	for itemColNum := 0; itemColNum < refiner.binNames.Len(); itemColNum++ {
		fn := refiner.binNames.Items[itemColNum]
		if strings.HasPrefix(fn, "udk_") {
			xType = storekeeper.UDK
		} else if strings.HasPrefix(fn, "sdk_") {
			xType = storekeeper.SDK
		} else if strings.HasPrefix(fn, "suk_") {
			xType = storekeeper.SUK
		} else if strings.HasPrefix(fn, "uuk_") {
			xType = storekeeper.UUK
		} else {
			continue
		}
		if strings.HasSuffix(fn, "_s.bin") {
			dType = storekeeper.SDT
		} else if strings.HasSuffix(fn, "_i.bin") {
			dType = storekeeper.IDT
		} else if strings.HasSuffix(fn, "_n.bin") {
			dType = storekeeper.NDT
		} else if strings.HasSuffix(fn, "_d.bin") {
			dType = storekeeper.DDT
		} else if strings.HasSuffix(fn, "_u.bin") {
			dType = storekeeper.UDT
		} else if strings.HasSuffix(fn, "_m.bin") {
			dType = storekeeper.MDT
		} else if strings.HasSuffix(fn, "_f.bin") {
			dType = storekeeper.FDT
		} else {
			continue
		}
		if err = interlacer.AddBinAsset(itemColNum, fn, dType, xType); err != nil {
			return err
		}
	}
	if err = interlacer.TsvToBin(); err != nil {
		return err
	}
	return interlacer.Save()
}

// OpenCowriterWithCatalog initializes a Cowriter and a catalog to write.
func (refiner *Refiner) OpenCowriterWithCatalog(assetName string) error {
	var err error
	if err = refiner.OpenCowriter(assetName); err != nil {
		return err
	}
	refiner.Cat = storekeeper.NewCataloguer()
	refiner.Cat.SetBinPathAndFileNames(refiner.Base.EntityCatalogPath(refiner.EntityNS), assetName)
	if err = refiner.Cat.CreateTmpFile(refiner.Base.EntityCatalogPath(refiner.EntityNS)); err != nil {
		return runner.ValidateErrors(err, refiner.CW.Close())
	}
	return nil
}

// OpenCowriter initializes a Cowriter.
func (refiner *Refiner) OpenCowriter(assetName string) error {
	var err error
	refiner.CW = cowriter.NewCowriter()
	if err = refiner.CW.Open(refiner.Base.CoWrAssetParams(refiner.EventNS, assetName)); err != nil {
		return runner.ValidateErrors(err, refiner.CW.Close())
	}
	return nil
}

// CloseCowriterWithCatalog saves the catalog, closes the Cowriter, sorts if necessary, and writes the bin properties.
func (refiner *Refiner) CloseCowriterWithCatalog(sortArgs string) error {
	var err error
	if refiner.Mgr.GetBitFlag(runner.MgrFlgNoDelete) {
		err = refiner.Cat.Save()
	} else {
		err = runner.ValidateErrors(refiner.Cat.Save(), refiner.Cat.RemoveTmpFile())
	}
	if err != nil {
		return runner.ValidateErrors(err, refiner.CW.Close())
	}
	return refiner.CloseCowriter(sortArgs)
}

// End completes the Refiner's process by copying the log file to the event's action folder.
func (refiner *Refiner) End(e error) {
	err := e
	if refiner.Base != nil && !refiner.Mgr.GetBitFlag(runner.MgrFlgHelp) && refiner.Mgr.GetBitFlag(runner.MgrFlgLogOK) {
		src := refiner.Mgr.GetConfigString(runner.LogFile)
		if src != "" && !strings.Contains(src, "/action/j") {
			log := fmt.Sprintf("%s_%s", runner.TimeTagNow(), path.Base(src))
			dst := path.Join(refiner.Base.EventActionPath(refiner.EventNS), log)
			if src != dst {
				refiner.Mgr.CopyLogFile(dst)
			}
		}
	}
	refiner.Mgr.End(err)
}

// CloseCowriter closes the Cowriter, sorts if necessary, and writes the bin properties.
// Attention, 3 sort options only: "-n", "-u", "-u -n", "none", else no sorting.
// Option "-u -n" sorts twice: first for unique tuples and then in numeric order.
func (refiner *Refiner) CloseCowriter(sortArgs string) error {
	var err error
	if err = refiner.CW.Close(); err != nil {
		return err
	}
	refiner.Mgr.Say(fmt.Sprintf("%s size: %d", refiner.CW.File.Name(), refiner.CW.BytesWritten()))
	if refiner.count > 0 {
		if sortArgs == "-n" {
			err = refiner.CW.Sort("-n")
		} else if sortArgs == "-u -n" {
			if err = refiner.CW.Sort("-u"); err == nil {
				err = refiner.CW.Sort("-n")
			}
		} else if sortArgs == "-u" {
			err = refiner.CW.Sort("-u")
		} else if sortArgs == "none" {
			err = refiner.CW.Sort("")
		}
		if err == nil {
			if refiner.Mgr.GetBitFlag(runner.MgrFlgNoDelete) {
				return refiner.InterlaceProperties()
			}
			return runner.ValidateErrors(refiner.InterlaceProperties(), refiner.CW.Remove())
		}
	}
	if refiner.Mgr.GetBitFlag(runner.MgrFlgNoDelete) {
		return err
	}
	return runner.ValidateErrors(err, refiner.CW.Remove())
}

func (refiner *Refiner) ResetCounter() {
	refiner.count = 0
}

func (refiner *Refiner) CountInc() {
	refiner.count++
}

func (refiner *Refiner) Count() int {
	return refiner.count
}

func (refiner *Refiner) SrcNames() []string {
	return refiner.srcNames
}

func (refiner *Refiner) UrlPaths() []string {
	return refiner.urlPaths
}

func (refiner *Refiner) PropertyEnumName(enumName string) string {
	return fmt.Sprintf("udk_%s_d.bin", enumName)
}

func (refiner *Refiner) EnumName(enumName string) string {
	return fmt.Sprintf("uuk_%s_s.bin", enumName)
}

func (refiner *Refiner) EnumFp(enumName string) string {
	return path.Join(refiner.Base.EntityEnumerationPath(refiner.EntityNS), refiner.EnumName(enumName))
}

// SrcInpFile retrieves the input file path associated with the given srcKeyName or falls back to a configuration value.
func (refiner *Refiner) SrcInpFile(srcKeyName string) string {
	if srcKeyName == "" {
		return ""
	}
	if unicode.IsUpper(rune(srcKeyName[0])) && !strings.Contains(srcKeyName, ".") {
		return refiner.Mgr.GetConfigString(srcKeyName)
	}
	return refiner.Mgr.DataPathInpFile(srcKeyName)
}

// OboParser parses the OBO file into a map of [OBO:ID] = term.
func (refiner *Refiner) OboParser(srcKeyName, ontologyName, separator string, obo map[string]string) error {
	var err error
	ska := runner.NewScanner(0)
	if err = ska.Start(refiner.SrcInpFile(srcKeyName)); err != nil {
		return runner.ValidateErrors(err, ska.Close())
	}
	var found bool
	var term, oid string
	records, row := 0, 0
	for row = 1; ska.Scan(); row++ {
		if strings.HasPrefix(ska.Text(), "[Term]") {
			records++
			oid = ""
		} else if strings.HasPrefix(ska.Text(), separator) {
			_, oid, found = strings.Cut(ska.Text(), ": ")
		} else if strings.HasPrefix(ska.Text(), "name: ") && oid != "" {
			_, term, found = strings.Cut(ska.Text(), ": ")
			if found && term != "" {
				obo[oid] = strings.TrimSpace(term)
				refiner.count++
			}
		}
	}
	refiner.ReportCounts(row, records, ontologyName+" Terms")
	refiner.ResetCounter()
	return runner.ValidateErrors(refiner.FlushAll(), ska.Close())
}

// MakeEnum creates and saves the bit flags enumerators.
func (refiner *Refiner) MakeEnum(enumName string, enumeration []string) error {
	if err := refiner.Enum.Create(enumeration); err != nil {
		return err
	}
	return refiner.Enum.Save(refiner.EnumFp(enumName))
}

// MakeStrArr creates and saves the string array using a container.
func (refiner *Refiner) MakeStrArr(binName string, arr []string) error {
	co := storekeeper.NewContainer(binName)
	return co.SaveStrArr(path.Join(refiner.Base.EntityPropertyPath(refiner.EntityNS), binName), arr)
}

// MakeNumArr creates and saves a numeric array into the specified binary file using a container.
func (refiner *Refiner) MakeNumArr(binName string, arr any) error {
	co := storekeeper.NewContainer(binName)
	/*
		switch arr.(type) {
		case []uint32:
			return co.SaveNumArr(path.Join(refiner.Base.EntityPropertyPath(refiner.EntityNS), binName), arr.([]uint32))
		case []uint64:
			return co.SaveNumArr(path.Join(refiner.Base.EntityPropertyPath(refiner.EntityNS), binName), arr.([]uint64))
		case []uint8:
			return co.SaveNumArr(path.Join(refiner.Base.EntityPropertyPath(refiner.EntityNS), binName), arr.([]uint8))
		case []int64:
			return co.SaveNumArr(path.Join(refiner.Base.EntityPropertyPath(refiner.EntityNS), binName), arr.([]int64))
		case []int32:
			return co.SaveNumArr(path.Join(refiner.Base.EntityPropertyPath(refiner.EntityNS), binName), arr.([]int32))
		case []float64:
			return co.SaveNumArr(path.Join(refiner.Base.EntityPropertyPath(refiner.EntityNS), binName), arr.([]float64))
		default:
			return fmt.Errorf("invalid type for numeric array: %T", arr)
		}
	*/
	return co.SaveNumArr(path.Join(refiner.Base.EntityPropertyPath(refiner.EntityNS), binName), arr)
}
