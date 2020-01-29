package fuse

type Match struct {
	// index of the Match within `list`
	item uint
	//  TODO: not sure what this is yet
	arrayIndex uint
	// actual value of the entity at index `item`
	value string // NOTE: this is a string for basic impl., but if adding map/dict support, it could be other types
	// score of this match
	score float32
	// array of indices that match
	matchedIndices [][2]uint

	/***** TODO: Properties related to map/dictionary support. *****/
	// key string
}

type SortFunc func(m1 Match, m2 Match) float32

type FuseOptions struct {
	// Approximately where in the text is the pattern expected to be found?
	Location uint
	// Determines how close the match must be to the fuzzy location (specified above).
	// An exact letter match which is 'distance' characters away from the fuzzy location
	// would score as a complete mismatch. A distance of '0' requires the match be at
	// the exact location specified, a threshold of '1000' would require a perfect match
	// to be within 800 characters of the fuzzy location to be found using a 0.8 threshold.
	Distance uint
	// At what point does the match algorithm give up. A threshold of '0.0' requires a perfect match
	// (of both letters and location), a threshold of '1.0' would match anything.
	Threshold float32
	// Machine word size
	MaxPatternLength uint
	// Indicates whether comparisons should be case sensitive.
	CaseSensitive bool
	// Regex used to separate words when searching. Only applicable when `tokenize` is `true`.
	/** tokenSeparator string **/
	// When true, the algorithm continues searching to the end of the input even if a perfect
	// match is found before the end of the same input.
	FindAllMatches bool
	// Minimum number of characters that must be matched before a result is considered a match
	MinMatchCharLength uint

	// Whether to sort the result list, by score
	ShouldSort bool

	// Default sort function
	SortFn SortFunc
	// When true, the search algorithm will search individual words **and** the full string,
	// computing the final score as a function of both. Note that when `tokenize` is `true`,
	// the `threshold`, `distance`, and `location` are inconsequential for individual tokens.
	Tokenize bool
	// When true, the result set will only include records that match all tokens. Will only work
	// if `tokenize` is also true.
	MatchAllTokens bool

	IncludeMatches bool
	IncludeScore   bool

	/***** TODO: Options related to map/dictionary support. *****/
	// The name of the identifier property. If specified, the returned result will be a list
	// of the items' dentifiers, otherwise it will be a list of the items.
	// id = null,
	// List of properties that will be searched. This also supports nested properties.
	// keys = [],
	// The get function to use when fetching an object's properties.
	// The default will search nested paths *ie foo.bar.baz*
	// getFn = deepValue,
}

/*******************************************************************/

// TODO: The resultMap is a key/value type that is built during the search to
//		  keep track of which results have already been added (prevents duplicates)
//		  It, and its subcomponents, are for internal use only.
//		 NOTE: I've discovered that this map is only required if the `list` is
//			   an object, and we're searching in multiple keys. In that case,
//			   the algorithm can yield duplicate matches if the searchText matches
//			   >1 key in a given item. Since I'm only supporting a flat list of
//			   strings for `list`, this map is not required.

/***** ResultMap (not a defined-type) *****/
//  The key is a string version of the entry index (within `Fuse.list`)
//	The value is a struct/object of type `Result`
//  {
//   '722' : { item: 722,  output: [ [Object] ] },
//   '1503': { item: 1503, output: [ [Object] ] },
//   '1514': { item: 1514, output: [ [Object] ] },
//   '1518': { item: 1518, output: [ [Object] ] },
//   '3891': { item: 3891, output: [ [Object] ] },
//   '5026': { item: 5026, output: [ [Object] ] },
//	  ...
//	}

/***** TODO: `Result` *****/
//		  `item` is the numeric version of the entry index (probably just to save us from having to use parseInt
//			 since keys MUST be Strings in JavaScript. However, in Go, we can use integers for the key, which
//			 means we can probably exclude the `item` field from this struct. Which, in turn, means that we can
//			 probably eliminate this entire struct `Result`, and just use an `Output` array as the value field of
//			 the resultMap
//		  `output` is an array of `Output` structs/objects, as shown below
// { item: 722,  output: [ [Object] ] },

/***** TODO: `Output` *****/
// 		key: '',
//     arrayIndex: -1, // NOTE: probably not necessary (map/dict related)
//     value: 'a spray of plum blossoms',
//     score: 0.5714285714285714,
//     matchedIndices: // NOTE: may not be necessary
//      [ [Array], [Array], [Array], [Array], [Array], [Array], [Array] ]

type FuseResult struct {
	item  int
	value string
	score float32
}

type Fuse struct {
	list    []string
	options FuseOptions
}
