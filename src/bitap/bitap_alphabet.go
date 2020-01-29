package bitap

import (
	"github.com/roderickjdunn/fusego/src/util"
)

// OPTIMIZATION: uint32 could potentially be smaller
func GetBitapAlphabet(pattern string) map[rune]uint32 {
	mask := make(map[rune]uint32)
	len := len(pattern)

	for i := 0; i < len; i += 1 {
		mask[[]rune(pattern)[i]] = 0
	}

	for i := 0; i < len; i += 1 {
		mask[[]rune(pattern)[i]] |= 1 << (len - i - 1)
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

	return accuracy + float32(proximity)/float32(distance)
}
