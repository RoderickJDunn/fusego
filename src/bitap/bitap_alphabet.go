package bitap

import (
	"fmt"

	"github.com/roderickjdunn/fusego/src/util"
)

// OPTIMIZATION: uint32 could potentially be smaller
func GetBitapAlphabet(pattern string) map[rune]uint32 {
	mask := make(map[rune]uint32)
	ptnLen := len(pattern)
	runePtn := []rune(pattern)
	runePtnLen := len(runePtn)

	if runePtnLen != ptnLen {
		fmt.Println("ERROR: Rune length differs! (", ptnLen, " vs. ", runePtnLen, ")")
		return mask
	}

	for i := 0; i < ptnLen; i++ {
		mask[runePtn[i]] = 0
	}

	for i := 0; i < ptnLen; i++ {
		mask[runePtn[i]] |= 1 << (ptnLen - i - 1)
	}

	return mask
}

// OPTIMIZATION: matchmask could be an actual mask, instead of an array
// OPTIMIZATION: we can probably remove the entire 'matchedIndices' functionality, to make it more light-weight
func getMatchedIndices(matchmask []uint8, minMatchCharLength uint) [][2]uint {

	var matchedIndices [][2]uint
	var start = -1
	var end = -1
	var i = 0

	for len := len(matchmask); i < len; i += 1 {
		var match = matchmask[i]
		if match == 1 && start == -1 {
			start = i
		} else if match == 0 && start != -1 {
			end = i - 1
			if uint((end-start)+1) >= minMatchCharLength {
				matchedIndices = append(matchedIndices, [2]uint{uint(start), uint(end)})
			}
			start = -1
		}
	}

	if matchmask[i-1] == 1 && uint(i-start) >= minMatchCharLength {
		matchedIndices = append(matchedIndices, [2]uint{uint(start), uint(i - 1)})
	}

	return matchedIndices
}

func getScore(pattern string, errors int, currentLocation int, expectedLocation int, distance uint) float32 {
	// fmt.Println("expectedLocation", expectedLocation)
	// fmt.Println("currentLocation", currentLocation)

	var accuracy float32 = float32(errors) / float32(len(pattern))
	var proximity uint = uint(util.AbsInt(expectedLocation - currentLocation))

	// fmt.Println("proximity", proximity)
	// fmt.Println("accuracy", accuracy)

	if distance == 0 {
		// Dodge divide by zero error.
		if proximity > 0 {
			return 1.0
		} else {
			return accuracy
		}
	}

	// NOTE: IMPORTANT. Accuracy == (error_count) / (pattern length)
	//
	//		*** Therefore, each error translates to a score increase of (1 / pattern_length) ***
	//
	//	    Since the overall score is the SUM of accuracy+(distance_from_exp_location),
	//		We can modify the accuracy *retroactively* from the user application.
	//		For example:
	//			Say we got a score of 0.5, for a pattern that is 25 characters long.
	//			The distance term doesn't matter, we can safely modify accuracy like so.
	//			If we got several small-word hits on this pattern, we can count the number of
	//			characters in those small-words hits (say 9 characters total), and add
	//
	//			sm_wd_acc_cost = 9 / 25 == 0.36
	//			ie. This term was given a 0.36 accuracy penalty due to its missing small words. Now
	//			we can remove that penalty from the original score.
	//
	//			modded_score = 0.5 - 0.36 == 0.14

	return accuracy + float32(proximity)/float32(distance)
}
