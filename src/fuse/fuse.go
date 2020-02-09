package fuse

import (
	"sort"
	"strings"

	"github.com/roderickjdunn/fusego/src/bitap"
)

// TODO: this is literally translated from JS for now, and doesn't work for GO sorts. Sorting functionality needs to be implemented
//		 properly by following this link: https://golang.org/pkg/sort/
func defaultSortFn(results []FuseResult, r1 int, r2 int) bool {
	return results[r1].Score < results[r2].Score
}

func DefaultOptions() *FuseOptions {
	defaultOptions := FuseOptions{
		Location:           0,
		Distance:           100,
		Threshold:          0.6,
		MaxPatternLength:   32,
		CaseSensitive:      false,
		FindAllMatches:     false,
		MinMatchCharLength: 1,
		ShouldSort:         false,
		SortFn:             defaultSortFn,
		Tokenize:           false,
		MatchAllTokens:     false,
		IncludeMatches:     false,
		IncludeScore:       false,
	}

	return &defaultOptions
}

func NewFuse(list []string, opts FuseOptions) Fuse {

	var fuse Fuse
	fuse.list = list
	fuse.options = opts

	return fuse
}

func prepareSearchers(pattern string, opts FuseOptions) (*bitap.Bitap, []*bitap.Bitap) {
	fullSearcher := bitap.NewBitap(pattern, opts.Location, opts.Distance, opts.Threshold,
		opts.MaxPatternLength, opts.CaseSensitive, opts.FindAllMatches, opts.MinMatchCharLength)

	var tokenSearchers []*bitap.Bitap

	if opts.Tokenize == true {
		tokens := strings.Fields(pattern) // TODO: tokenization is limited to using WhiteSpace as the delimiter for now

		// OPTIMIZATION: ? seems inefficient in 2 ways. 1) We are re-assigning the result to `tokenSearchers` on every iteration,
		//				 and 2) we are accessing `opts` properties on every iteration
		len := len(tokens)
		for i := 0; i < len; i++ {
			toAdd := bitap.NewBitap(tokens[i], opts.Location, opts.Distance, opts.Threshold,
				opts.MaxPatternLength, opts.CaseSensitive, opts.FindAllMatches, opts.MinMatchCharLength)
			tokenSearchers = append(tokenSearchers, toAdd)
		}
	}

	return fullSearcher, tokenSearchers

}

func _search(fuse Fuse, tokenSearchers []*bitap.Bitap, fullSearcher *bitap.Bitap) []FuseResult {
	list := fuse.list
	var results []FuseResult

	len := len(list)
	// Iterate over every item
	for i := 0; i < len; i++ {
		results = _analyze(fuse, list[i], i, tokenSearchers, fullSearcher, results)

		// _analyze({
		// 	key: '',
		// 	value: list[i],
		// 	record: i,
		// 	index: i
		// }, {
		// 	results,
		// 	tokenSearchers,
		// 	fullSearcher
		// })
	}

	return results

}

func _analyze(fuse Fuse, value string, index int, tkSeachers []*bitap.Bitap, fullSearcher *bitap.Bitap, results []FuseResult) []FuseResult {
	// Check if the texvaluet can be searched
	if value == "" {
		return results
	}

	exists := false
	var averageScore float32 = -1
	numTextMatches := 0

	isMatchFullS := false
	var scoreFullS float32 = 1.0
	// var matchedIndicesFullS [][2]uint

	// if (typeof value === 'string') {
	// this._log(`\nKey: ${key === '' ? '-' : key}`)

	// NOTE: using _ instead of matchedIndicesFullS, since its not needed, and otherwise throws a 'not used' error
	isMatchFullS, scoreFullS, _ = bitap.Search(value, fullSearcher)
	// fmt.Println("value: ", value)
	// fmt.Println("isMatchFullS: ", isMatchFullS)
	// fmt.Println("scoreFullS: ", scoreFullS)

	if fuse.options.Tokenize == true {
		words := strings.Fields(value)
		wordCnt := len(words)
		var scores []float32

		tkSearchersLen := len(tkSeachers)

		for i := 0; i < tkSearchersLen; i++ {
			tokenSearcher := tkSeachers[i]

			// this._log(`\nPattern: "${tokenSearcher.pattern}"`)

			hasMatchInText := false

			for j := 0; j < wordCnt; j++ {
				word := words[j]

				// NOTE: using _ instead of matchedIndicesTk, since its not needed, and otherwise throws a 'not used' error
				isMatchTkS, scoreTkS, _ := bitap.Search(word, tokenSearcher)
				// tokenSearchResult = bitap.Search(word, tokenSearcher)
				// let obj = {}
				if isMatchTkS == true {
					//   obj[word] = tokenSearchResult.Score
					exists = true
					hasMatchInText = true
					scores = append(scores, scoreTkS)
				} else {
					//   obj[word] = 1
					if !fuse.options.MatchAllTokens {
						scores = append(scores, 1)
					}
				}
				// this._log(`Token: "${word}", score: ${obj[word]}`)
				// tokenScores.push(obj)
			}

			if hasMatchInText {
				numTextMatches++
			}
		}

		averageScore = scores[0]
		scoresLen := len(scores)
		for i := 1; i < scoresLen; i++ {
			averageScore += scores[i]
		}
		averageScore = averageScore / float32(scoresLen)

		// fmt.Println("averageScore (tk): ", averageScore)
		// this._log('Token score average:', averageScore)
	}

	var finalScore float32 = scoreFullS
	if averageScore > -1 {
		finalScore = (finalScore + averageScore) / 2
	}

	// this._log('Score average:', finalScore)

	// Translation: checkTextMatches == true if...
	//    1) .tokenize OR .matchAllTokens is false
	//    or
	//    2) numTextMatches >= tokenSearchers.length
	// => False only if
	//    1) .tokenize AND .matchAllTokens are true, AND numTextMatches < tokenSearchers.length

	checkTextMatches := true
	if fuse.options.Tokenize && fuse.options.FindAllMatches && numTextMatches < len(tkSeachers) {
		checkTextMatches = false
	}

	// this._log(`\nCheck Matches: ${checkTextMatches}`)

	// If a match is found, add the item to <rawResults>, including its score.
	// EXPLANATION: - `exists` is flag that indicates we found a match using tokenSearch
	//              - `mainSearchResult.isMatch` indicates that a match was found using the fullSearch
	//              If either of these is true, then a match was found, and we almost always want to
	//              add the result if a match was found.
	//              - `checkTextMatches` seems to be somewhat poorly-named. Its really just a flag
	//                that indicates whether a tokenSearch with matchAllTokens==true was successful.
	//                If there was no token search at all, or there was but matchAllTokens is false, or all tokens were matched
	//                with matchAllTokens==true, then this flag is true.
	//                In other words, the only time this flag is false, is when we did a matchAllTokens
	//                tokenSearch, and not all tokens were found.
	if (exists || isMatchFullS) && checkTextMatches {
		// Check if the item already exists in our results

		// NOTE: IMPORTANT: It seems that the only reason for needing the 'existingResult' map is because
		//                  when the 'list' is a map/object, and we're searching for text in multiple keys,
		//                  then a given item could get matched on >1 key, resulting in duplicate results.
		//                  This could never happen when the list is a flat list of strings, so we can
		//                  greatly simplify this logic and structs for Version 1.
		// Add it to the raw result list
		result := FuseResult{
			Item:  index,
			Value: value,
			Score: finalScore,
			// matchedIndices: mainSearchResult.matchedIndices
		}
		return append(results, result)
	} else {
		return results // return unmodified results
	}
}

func FuseSearch(fuse Fuse, pattern string) []FuseResult {

	fullSearcher, tokenSearchers := prepareSearchers(pattern, fuse.options)
	// fmt.Println("fullSearcher", fullSearcher)
	// fmt.Println("tokenSearchers", tokenSearchers)

	results := _search(fuse, tokenSearchers, fullSearcher)

	// sort.Sort(ByScore(results))

	sort.Slice(results, func(i, j int) bool {
		return fuse.options.SortFn(results, i, j)
	})

	return results
	// resLen := len(results)

	// for i := 0; i < resLen; i++ {
	// 	fmt.Println(results[i])
	// }
	//   this._computeScore(weights, results)

	//   if (this.options.shouldSort) {
	// 	this._sort(results)
	//   }

	//   if (opts.limit && typeof opts.limit === 'number') {
	// 	results = results.slice(0, opts.limit)
	//   }

	//   return this._format(results)

	// DEV: returning empty matches
	// var matches []Match
	// return matches
}
