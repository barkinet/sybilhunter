// Looks for relays manipulating their fingerprints.

package main

import (
	"fmt"
	"os"

	tor "git.torproject.org/user/phw/zoossh.git"
)

type ValueSorter struct {
	// IP addresses in string format.
	keys []string
	// Amount of unique fingerprints.
	vals []int
}

// Implement the sort interface (1/3).
func (vs ValueSorter) Len() int {
	return len(vs.keys)
}

// Implement the sort interface (2/3).
func (vs ValueSorter) Swap(i int, j int) {
	vs.keys[i], vs.keys[j] = vs.keys[j], vs.keys[i]
	vs.vals[i], vs.vals[j] = vs.vals[j], vs.vals[i]
}

// Implement the sort interface (3/3).
func (vs ValueSorter) Less(i int, j int) bool {
	return vs.vals[i] < vs.vals[j]
}

// Used to count how often a given fingerprint was observed.
type FprStats map[string]int

// Go does not like net.IP as a map key.  So we use an IP address's string
// representation instead.
var FprAnalysis map[string]FprStats = map[string]FprStats{}

// countFingerprints updates the fingerprint statistics with the given
// fingerprint and address.
func countFingerprints(fpr string, address string) {

	fprStats, ok := FprAnalysis[address]
	if ok {
		_, ok := fprStats[fpr]
		if ok {
			// Fingerprint already present for address: update counter.
			fprStats[fpr] += 1
		} else {
			// Fingerprint new: add it to the map.
			fprStats[fpr] = 1
		}
	} else {
		FprAnalysis[address] = FprStats{fpr: 1}
	}
}

// AnalyseFingerprints parses the given file and then determines fingerprint
// statistics.  If something fails, an error is returned.
func AnalyseFingerprints(path string, info os.FileInfo, err error) error {

	if _, err = os.Stat(path); err != nil {
		fmt.Errorf("File \"%s\" does not exist.", path)
		return nil
	}

	if info.IsDir() {
		return nil
	}

	objects, err := tor.ParseUnknownFile(path)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	// Use a type switch to figure out what data we are dealing with and then
	// determine statistics.
	switch v := objects.(type) {
	case *tor.Consensus:
		for fpr, getAddr := range v.RouterStatuses {
			countFingerprints(fpr, getAddr().Address.String())
		}
	case *tor.RouterDescriptors:
		for fpr, getAddr := range v.RouterDescriptors {
			countFingerprints(fpr, getAddr().Address.String())
		}
	default:
		return fmt.Errorf("Data structure not implemented yet.")
	}

	return nil
}