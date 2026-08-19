package spellcheck

// rules_commas.go is the comma, which is the largest thing this plugin had
// nothing to say about.
//
// The mechanical half is in rules_punctuation.go and always was: a comma
// before a full stop, a missing space after one, the fixed transitions that
// always take one. What was missing is the half that depends on the sentence
// -- whether two clauses are one sentence or two, whether a clause defines the
// thing before it or merely adds to it, whether a name is being addressed or
// described. None of those can be settled by looking at the words, and every
// one of them is settled differently by different good writers.
//
// So they are checks. This is the case the tier was built for and the reason
// it is worth having at all: an error rule would have to be right about
// somebody's sentence structure, which no expression over words can be, and a
// style suggestion would be claiming the writing could read better when the
// claim is that a mark may be missing. A check asks, offers the comma, and
// takes "Correct Usage" for an answer -- which for punctuation, where house
// style is a real thing people hold to, is the answer that matters most.
//
// Category "Punctuation" rather than "Possible confusion": these sit beside
// the mechanical comma rules in the reader's head, and the heading is what
// says so. The severity, not the category, is what puts them on the
// Suggestions and Checks page.
//
// Two shapes are deliberately absent. A comma before "and" joining two
// clauses cannot be told from a compound subject -- "Sam and I went" needs no
// comma and looks identical to a regex. And the serial comma is a house style
// with two correct answers, which is not a question worth asking on every list
// somebody writes.

var commaCheckRules = []GrammarRule{
	{
		// A comma splice: two sentences with a comma between them. The second
		// clause is spotted by a pronoun and a verb, which is what a clause
		// starts with when it is a clause and not a list item.
		Pattern: `\b(\w{3,}),\s+([iI]t|[hH]e|[sS]he|[tT]hey|[wW]e|[yY]ou|[iI])\s+` +
			`(is|was|were|are|has|had|have|will|would|can|could|did|does|` +
			`went|got|took|made|knew|thought|\w+ed)\b`,
		Antipatterns: []string{
			// Reported speech, where the comma is exactly right.
			`\b([sS]aid|[sS]ays|[aA]sked|[tT]old|[rR]eplied|[aA]dded|[nN]oted|` +
				`[eE]xplained|[wW]rote|[wW]hispered|[sS]houted),\s+\w+\s+\w+`,
			// A subordinate clause first: "When I got home, it was late" is
			// one sentence, and the comma is what makes it one.
			`\b([iI]f|[wW]hen|[wW]hile|[aA]fter|[bB]efore|[aA]lthough|[tT]hough|` +
				`[bB]ecause|[sS]ince|[uU]nless|[uU]ntil|[wW]henever|[oO]nce|` +
				`[wW]hether)\b[^,.!?]*,\s+\w+\s+\w+`,
			// A single word opening the sentence -- "Instead, it was fine",
			// "Yesterday, we shipped" -- which is an adverb and not a clause.
			`(^|[.!?]\s|\n)\w+,\s+\w+\s+\w+`,
			// The same job done by a phrase rather than a word. These have no
			// verb in them, so what follows the comma is the sentence's only
			// clause and nothing is being spliced. Listed rather than matched
			// as "any few words", which would take the first clause of a real
			// splice with it -- "I went home, it was late" is three words and
			// a comma too, and the difference is a verb this cannot see.
			`(^|[.!?]\s|\n)(Other than that|In any case|As for that|` +
				`After all|By the way|Of course|In fact|For example|At least|` +
				`On the whole|Even so|That said|In short|Apart from that|` +
				`In the end|At the same time|On balance|If anything|` +
				`All the same|To be fair|In the meantime|Either way),\s+\w+\s+\w+`,
		},
		Flags: []string{
			"I went home, it was late",
			"the release slipped, we shipped anyway",
		},
		Leaves: []string{
			"she said, it was late",
			"when I got home, it was late",
			"Instead, it was fine",
			"Other than that, it went fine",
			"I went home; it was late",
		},
		Message:  "Two sentences joined by a comma?",
		Suggest:  "$1; $2 $3",
		Category: "Punctuation",
		Severity: SeverityCheck,
		Explanation: "Two complete sentences take a semicolon or a full stop " +
			"between them rather than a comma. Written deliberately, a comma " +
			"between two short clauses is a style some writers keep.",
	},
	{
		// "Which" with no comma in front of it. In the distinction most style
		// guides draw, a clause that merely adds something takes a comma and a
		// clause that defines the thing before it takes "that" -- and British
		// writing has used a bare "which" for both for centuries, which is
		// exactly why this asks rather than corrects.
		Pattern: `\b(\w{3,})\s+which\b`,
		Antipatterns: []string{
			// A preposition takes "which" and no comma: "the way in which".
			`\b(in|on|at|by|for|to|with|from|of|through|during|after|before|` +
				`about|into|upon|within|against|toward|towards)\s+which\b`,
			// An indirect question: "I don't know which to pick", "deciding
			// which way to go". The -ing forms matter as much as the plain
			// ones and were missing at first, which is how the corpus caught
			// this rule on a perfectly good sentence.
			`\b([kK]now|[kK]nows|[kK]new|[kK]nowing|[sS]ee|[sS]aw|[sS]eeing|` +
				`[tT]ell|[tT]elling|[dD]ecide|[dD]ecided|[dD]eciding|` +
				`[cC]hoose|[cC]hose|[cC]hoosing|[aA]sk|[aA]sked|[aA]sking|` +
				`[wW]onder|[wW]ondering|[cC]heck|[cC]hecking|[sS]ure|` +
				`[mM]atter|[dD]epends|[dD]epending|[gG]uess|[gG]uessing|` +
				`[rR]emember|[rR]emembering|[sS]how|[sS]hows|[sS]howing)\s+which\b`,
		},
		Flags: []string{
			"the release which slipped was the big one",
			"a rule which fires on correct writing",
		},
		Leaves: []string{
			"the release, which slipped, was the big one",
			"the way in which it was written",
			"I don't know which to pick",
			"we weighed it before deciding which way to go",
		},
		Message:  "Comma before \"which\"?",
		Suggest:  "$1, which",
		Category: "Punctuation",
		Severity: SeverityCheck,
		Explanation: "A clause that adds something takes a comma before " +
			"\"which\"; one that says which thing you mean is usually " +
			"written with \"that\" and no comma. British writing often uses " +
			"\"which\" for both.",
	},
	{
		// A name being addressed is set off by a comma: "Thanks Ken" against
		// "Thanks, Ken". Common enough in a chat client to be worth asking
		// about, and impossible to settle -- the plugin has no way to know
		// whether the capitalised word is a person or a thing.
		Pattern: `(^|[.!?]\s|\n)(Hi|Hey|Hello|Thanks|Thank you|Sorry|Welcome|` +
			`Goodbye|Bye|Congratulations|Good morning|Good afternoon|` +
			`Good evening)\s+([A-Z][a-z]{2,})\b`,
		Antipatterns: []string{
			// The set phrases where the capitalised word is not a person.
			`(Hello\s+World|Thanks\s+(God|Goodness|Heaven|Heavens)|` +
				`Sorry\s+(State|Sight|Excuse)|Welcome\s+(Pack|Page|Screen|` +
				`Email|Message|Guide|Bonus)|Good\s+(Friday|Samaritan))`,
		},
		Flags: []string{
			"Thanks Ken for the review",
			"Hi Sarah",
			"we shipped it. Thanks Ken",
		},
		Leaves: []string{
			"Thanks, Ken for the review",
			"Hello World is the first program",
			"Welcome Pack arrived today",
		},
		Message:  "Comma before a name?",
		Suggest:  "$1$2, $3",
		Category: "Punctuation",
		Severity: SeverityCheck,
		Explanation: "A name you are speaking to is separated by a comma: " +
			"\"Thanks, Ken\". Without it the name reads as part of what is " +
			"being thanked, which is occasionally what was meant.",
	},
	{
		// An introductory clause takes a comma before the main one. Kept to a
		// short clause -- two to six lowercase words -- because a long one is
		// where the comma is least likely to have been left out on purpose
		// and most likely to be somewhere this cannot see.
		Pattern: `(^|[.!?]\s|\n)([Ww]hen|[Ii]f|[Aa]fter|[Bb]efore|[Aa]lthough|` +
			`[Ww]hile|[Ss]ince|[Bb]ecause|[Uu]nless|[Uu]ntil|[Oo]nce)\s+` +
			`([a-z]+(?:\s+[a-z]+){1,5})\s+(we|I|he|she|it|they|you)\s+` +
			`(will|would|can|could|should|must|might|is|was|are|were|have|` +
			`has|had|do|does|did|went|got|need|needs)\b`,
		Antipatterns: []string{
			// An inverted question opening with the same word: "Since when do
			// we have to ask?" is not a clause waiting for a comma.
			`(^|[.!?]\s|\n)([Ss]ince|[Ww]hen|[Uu]ntil)\s+` +
				`(when|how|why|what|where)\s+(do|does|did|can|could|will|` +
				`would|should|have|has)\s+(we|I|you|he|she|it|they)\s+\w+`,
		},
		Flags: []string{
			"When the release shipped we went home",
			"If the build fails we should retry",
		},
		Leaves: []string{
			"When the release shipped, we went home",
			"If it fails, we retry",
			"Since when do we have to ask for that",
		},
		Message:  "Comma after the opening clause?",
		Suggest:  "$1$2 $3, $4 $5",
		Category: "Punctuation",
		Severity: SeverityCheck,
		Explanation: "A clause that opens a sentence is usually separated " +
			"from the main one by a comma. A very short opener is often left " +
			"unpunctuated on purpose.",
	},
	{
		// The same shape as the "but" rule in rules_punctuation.go, which is
		// a suggestion, applied to "so" -- where the reading is less settled,
		// since "so" also introduces a purpose clause that takes no comma.
		Pattern: `\b(\w{3,})\s+so\s+(I|we|you|he|she|it|they)\s+` +
			`(will|would|can|could|should|must|might|is|was|are|were|have|` +
			`has|had|did|do|went|got|took|made|\w+ed)\b`,
		Antipatterns: []string{
			// A purpose clause: "held it open so we could pass". Written from
			// the word before "so", since an exception suppresses only where
			// it contains the match and the match starts there.
			`\b\w{3,}\s+so\s+(I|we|you|he|she|it|they)\s+` +
				`(can|could|would|will|might)\b`,
		},
		Flags: []string{
			"the build failed so we shipped late",
			"nobody replied so I did it myself",
		},
		Leaves: []string{
			"the build failed, so we shipped late",
			"he held the door so we could pass",
		},
		Message:  "Comma before \"so\"?",
		Suggest:  "$1, so $2 $3",
		Category: "Punctuation",
		Severity: SeverityCheck,
		Explanation: "\"So\" joining two complete sentences takes a comma " +
			"before it. \"So\" meaning \"in order that\" does not, and the " +
			"two look alike.",
	},
}
