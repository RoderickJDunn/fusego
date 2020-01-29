package bitap

import (
	"math"
	"strings"

	"github.com/roderickjdunn/fusego/src/util"
)

type Bitap struct {
	pattern string
	// Approximately where in the text is the pattern expected to be found?
	location uint
	// Determines how close the match must be to the fuzzy location (specified above).
	// An exact letter match which is 'distance' characters away from the fuzzy location
	// would score as a complete mismatch. A distance of '0' requires the match be at
	// the exact location specified, a threshold of '1000' would require a perfect match
	// to be within 800 characters of the fuzzy location to be found using a 0.8 threshold.
	distance uint
	// At what point does the match algorithm give up. A threshold of '0.0' requires a perfect match
	// (of both letters and location), a threshold of '1.0' would match anything.
	threshold float32
	// Machine word size
	maxPatternLength uint
	// Indicates whether comparisons should be case sensitive.
	isCaseSensitive bool
	// Regex used to separate words when searching. Only applicable when `tokenize` is `true`.
	/** tokenSeparator string **/
	// When true, the algorithm continues searching to the end of the input even if a perfect
	// match is found before the end of the same input.
	findAllMatches bool
	// Minimum number of characters that must be matched before a result is considered a match
	minMatchCharLength uint
	/**** private fields ****/
	//
	patternAlphabet map[rune]uint32
}

func NewBitap(pattern string,
	location uint, distance uint, threshold float32, maxPatternLength uint, isCaseSensitive bool,
	findAllMatches bool, minMatchCharLength uint) *Bitap {

	// OPTIMIZATION: just assume its lowercase
	if isCaseSensitive == false {
		pattern = strings.ToLower(pattern)
	}

	var alphabet map[rune]uint32
	if len(pattern) <= int(maxPatternLength) {
		alphabet = GetBitapAlphabet(pattern)
	}

	bt := Bitap{pattern, location, distance, threshold, maxPatternLength, isCaseSensitive, findAllMatches, minMatchCharLength, alphabet}
	return &bt
}

func BitapSearch(text string, bitap Bitap) (isMatch bool, score float32, matchedIndices [][2]uint) {

	expectedLocation := int(bitap.location)
	// Set starting location at beginning text and initialize the alphabet.
	textLen := len(text)
	// Highest score beyond which we give up.
	currentThreshold := bitap.threshold
	// Is there a nearby exact match? (speedup)
	var bestLocation int = util.IndexFrom(text, bitap.pattern, expectedLocation)

	patternLen := len(bitap.pattern)
	// // fmt.Println("patternLen", patternLen)

	// a mask of the matches
	matchMask := make([]uint8, textLen)

	for i := range matchMask {
		matchMask[i] = 0
	}
	// fmt.Println("bestLocation", bestLocation)

	if bestLocation != -1 {
		score := getScore(bitap.pattern, 0, bestLocation, int(expectedLocation), bitap.distance)
		// fmt.Println("score", score)
		currentThreshold = float32(math.Min(float64(score), float64(currentThreshold)))

		// What about in the other direction? (speed up)
		bestLocation = util.LastIndexFrom(text, bitap.pattern, expectedLocation+patternLen)

		if bestLocation != -1 {
			score = getScore(bitap.pattern, 0, bestLocation, int(expectedLocation), bitap.distance)
			currentThreshold = float32(math.Min(float64(score), float64(currentThreshold)))
			// fmt.Println("score", score)
		}
	}

	// Reset the best location
	bestLocation = -1

	var lastBitArr []uint

	finalScore := 1.0
	binMax := patternLen + textLen

	var mask uint
	if patternLen <= 31 {
		mask = 1 << (patternLen - 1)
	} else {
		mask = 1 << 30
	}

	// // fmt.Println("mask", mask)

	for i := 0; i < patternLen; i++ {
		// // fmt.Println("i", i)
		// Scan for the best match; each iteration allows for one more error.
		// Run a binary search to determine how far from the match location we can stray
		// at this error level.
		binMin := 0
		binMid := binMax

		for binMin < binMid {
			score := getScore(bitap.pattern, i, int(expectedLocation+binMid), int(expectedLocation), bitap.distance)

			// // fmt.Println("currentThreshold", currentThreshold)
			// fmt.Println("score", score)

			if score <= currentThreshold {
				// // fmt.Println("score <= currentThreshold")
				binMin = binMid
			} else {
				// // fmt.Println("score > currentThreshold")
				binMax = binMid
			}

			binMid = int(math.Floor((float64(binMax-binMin) / 2.0) + float64(binMin)))
			// fmt.Println("mid", binMid)
			// // fmt.Println("max", binMax)
			// // fmt.Println("min", binMin)
			// // fmt.Println("---")
		}

		// Use the result from this iteration as the maximum for the next.
		binMax = binMid

		start := util.MaxInt(1, int(expectedLocation)-binMid+1)
		// fmt.Println("binMid", binMid)
		// fmt.Println("expectedLocation", expectedLocation)
		// // fmt.Println("binMid", binMid)
		// // fmt.Println("expectedLocation", expectedLocation)
		// // fmt.Println("start", start)

		var finish int
		if bitap.findAllMatches {
			finish = textLen
			// // fmt.Println("finish (a)", finish)
		} else {
			finish = util.MinInt(expectedLocation+binMid, textLen) + patternLen
			// // fmt.Println("finish (b)", finish)
		}

		// Initialize the bit array
		bitArr := make([]uint, finish+2)

		bitArr[finish+1] = (1 << i) - 1
		// fmt.Println("start", start)
		// fmt.Println("finish", finish)

		for j := finish; j >= start; j-- {
			currentLocation := j - 1
			// fmt.Println("curr location: ", currentLocation)
			// fmt.Println("text: ", text)
			var charMatch uint32
			foundinAlpha := false
			if currentLocation < len(text) {
				charMatch, foundinAlpha = bitap.patternAlphabet[rune(text[currentLocation])]
			}
			// if !ok {
			// 	// // fmt.Println("Failed to find value in patternAlphabet: ", text[currentLocation])
			// }

			if foundinAlpha {
				matchMask[currentLocation] = 1
			}

			// First pass: exact match
			bitArr[j] = ((bitArr[j+1] << 1) | 1) & uint(charMatch)

			// Subsequent passes: fuzzy match
			if i != 0 {
				bitArr[j] |= (((lastBitArr[j+1] | lastBitArr[j]) << 1) | 1) | lastBitArr[j+1]
			}

			// fmt.Println("bitArr[j]", bitArr[j])
			// fmt.Println("mask", mask)

			if bitArr[j]&mask > 0 {
				finalScore = float64(getScore(bitap.pattern, i, int(currentLocation), int(expectedLocation), bitap.distance))
				// fmt.Println("finalScore", finalScore)

				// This match will almost certainly be better than any existing match.
				// But check anyway.
				if finalScore <= float64(currentThreshold) {
					// Indeed it is
					currentThreshold = float32(finalScore)
					bestLocation = int(currentLocation)

					// Already passed `loc`, downhill from here on in.
					if bestLocation <= expectedLocation {
						break
					}

					// When passing `bestLocation`, don't exceed our current distance from `expectedLocation`.
					start = util.MaxInt(1, 2*expectedLocation-bestLocation)
				}
			}
		}

		// No hope for a (better) match at greater error levels.
		score := getScore(bitap.pattern, i+1, expectedLocation, expectedLocation, bitap.distance)

		// fmt.Println("score", score, finalScore)

		if score > currentThreshold {
			break
		}

		lastBitArr = bitArr
	}

	//console.log('FINAL SCORE', finalScore)

	if finalScore == 0 {
		finalScore = 0.001
	}

	// Count exact matches (those with a score of 0) to be "almost" exact
	return bestLocation >= 0,
		float32(finalScore),
		getMatchedIndices(matchMask, bitap.minMatchCharLength)

}

// import "regexp"
/* NOTE: regex search is only used by fuse when pattern length is greater than
   the machine word length. Not going to implement for prototype.
*/
// var SPECIAL_CHARS_REGEX string = `/[\-\[\]\/\{\}\(\)\*\+\?\.\\\^\$\|]/g`

// // module.exports = (text, pattern, tokenSeparator = / +/g) => {
// func regexSearch(text string, pattern string) {
// 	// NOTE: Not including tokenSeparator arg. Just assume its a space.

// 	// escape special characters, then replace `tokenSeparator` (space) with pipe ('|')
//   let regex = new RegExp(pattern.replace(SPECIAL_CHARS_REGEX, '\\$&').replace(tokenSeparator, '|'))
//   let matches = text.match(regex)
//   let isMatch = !!matches
//   let matchedIndices = []

//   if (isMatch) {
//     for (let i = 0, matchesLen = matches.length; i < matchesLen; i += 1) {
//       let match = matches[i]
//       matchedIndices.push([text.indexOf(match), match.length - 1])
//     }
//   }

//   return {
//     // TODO: revisit this score
//     score: isMatch ? 0.5 : 1,
//     isMatch,
//     matchedIndices
//   }
// }

func Search(text string, bitap *Bitap) (isMatch bool, score float32, matchedIndices [][2]uint) {

	// OPTIMIZATION: Assume text is lowerCase
	if !bitap.isCaseSensitive {
		text = strings.ToLower(text)
	}

	// Exact match
	if bitap.pattern == text {
		matchedIndices := [][2]uint{
			{0, uint(len(text) - 1)},
		}
		return true, 0, matchedIndices
	}

	// TODO: When pattern length is greater than the machine word length, just do a a regex comparison
	// const { maxPatternLength, tokenSeparator } = this.options
	// if (this.pattern.length > maxPatternLength) {
	//   return bitapRegexSearch(text, this.pattern, tokenSeparator)
	// }
	// for now just log error and exit if greater than word length
	if uint(len(bitap.pattern)) > bitap.maxPatternLength {
		// fmt.Println("Pattern is too long for bitap search. TODO: add regex search!")
		matchedIndices := [][2]uint{}
		return false, 0, matchedIndices
	}

	// Otherwise, use Bitap algorithm
	return BitapSearch(text, *bitap)
}
