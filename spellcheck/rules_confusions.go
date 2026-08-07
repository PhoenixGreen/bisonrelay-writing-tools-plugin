package spellcheck

// rules_confusions.go holds the word pairs people mix up: their/there,
// then/than, lose/loose, affect/effect.
//
// These are the checks a grammar tool is judged on and the ones it most often
// gets wrong. LanguageTool decides the general case with ngram frequency data
// -- given "I have more money then you", it knows "money than you" is the
// commoner trigram -- which is gigabytes and out of the question here.
//
// What is possible without any of that is the decidable slice. Each pair has
// positions where only one of the two words can be right whatever the
// sentence is about: "than" cannot follow "and", a possessive cannot precede
// a verb, "apart of" is never a phrase. The rules below sit only in those
// positions and say nothing at all about the rest, which is why the list is
// so much shorter than the number of pairs it covers.
//
// The temptation throughout is to widen a rule by one more word. The corpus
// in spellcheck_test.go exists because that temptation was acted on several
// times, and every one of them fired on correct writing.

// confusionExplanation is shared by rules for one pair, so a reader gets the
// same account of the difference wherever it is caught.
const (
	explainThenThan = "\"Than\" compares two things (\"bigger than\"). " +
		"\"Then\" is about time or sequence (\"and then we left\")."
	explainLoseLoose = "\"Lose\" is the verb -- to mislay or be beaten. " +
		"\"Loose\" is the adjective, the opposite of tight."
	explainAffectEffect = "\"Affect\" is almost always the verb, meaning to " +
		"influence. \"Effect\" is almost always the noun, the result."
	explainYourYoure = "\"Your\" shows possession (\"your wallet\"). " +
		"\"You're\" is short for \"you are\"."
	explainToToo = "\"Too\" means excessively or also. \"To\" is the " +
		"preposition, as in \"to the shop\"."
	explainTheirTheyre = "\"Their\" shows possession (\"their wallet\"). " +
		"\"They're\" is short for \"they are\"."
	explainWhoseWhos = "\"Whose\" asks who something belongs to. " +
		"\"Who's\" is short for \"who is\" or \"who has\"."
	explainParticiple = "After \"have\", \"has\" or \"had\", a verb takes " +
		"its past participle -- \"have gone\", not \"have went\"."
)

// confusionRules are checks for the wrong word of a confusable pair, in
// positions where the right one is not a matter of opinion.
var confusionRules = []GrammarRule{
	// --- then / than ---
	{
		// A comparative is the one position where "than" is forced. Written
		// as an explicit list rather than `\w+er`, which would sweep up
		// "never", "however", "either" and "other" -- all of which precede
		// "then" perfectly well.
		Pattern: `\b(more|less|better|worse|bigger|smaller|larger|higher|` +
			`lower|faster|slower|greater|fewer|older|younger|longer|shorter|` +
			`cheaper|easier|harder|rather|other)\s+then\b`,
		Message:     "Should be \"$1 than\"",
		Suggest:     "$1 than",
		Category:    "Confused words",
		Explanation: explainThenThan,
	},
	{
		// The mirror. A conjunction cannot be followed by a comparison.
		Pattern:     `\b(and|but|so|back|since|right|only|way)\s+than\b`,
		Message:     "Should be \"$1 then\"",
		Suggest:     "$1 then",
		Category:    "Confused words",
		Explanation: explainThenThan,
	},

	// --- lose / loose ---
	{
		// "Loosing" is not a word anyone means: the verb "to loose" exists
		// but its participle is vanishingly rare beside the typo.
		Pattern:     `\bloosing\b`,
		Message:     "Should be \"losing\"",
		Suggest:     "losing",
		Category:    "Confused words",
		Explanation: explainLoseLoose,
	},
	{
		// An adjective cannot take an object, so anything that looks like one
		// after "loose" means the verb was meant.
		Pattern: `\bloose\s+(the|a|an|my|your|his|her|our|their|it|them|` +
			`money|weight|control|track|interest|everything|access)\b`,
		Message:     "Should be \"lose $1\"",
		Suggest:     "lose $1",
		Category:    "Confused words",
		Explanation: explainLoseLoose,
	},
	{
		// A modal or "to" forces the verb.
		Pattern:     `\b(will|would|can|could|might|may|to|don't|didn't|doesn't)\s+loose\b`,
		Message:     "Should be \"$1 lose\"",
		Suggest:     "$1 lose",
		Category:    "Confused words",
		Explanation: explainLoseLoose,
	},

	// --- affect / effect ---
	{
		// A determiner forces a noun.
		Pattern:     `\b(a|an|the|this|that|any|no|side|net|little|great|whose)\s+affect\b`,
		Message:     "Should be \"$1 effect\"",
		Suggest:     "$1 effect",
		Category:    "Confused words",
		Explanation: explainAffectEffect,
	},
	{
		Pattern:     `\b(the|these|those|its|side|any|no)\s+affects\b`,
		Message:     "Should be \"$1 effects\"",
		Suggest:     "$1 effects",
		Category:    "Confused words",
		Explanation: explainAffectEffect,
	},
	{
		// A modal forces the verb. "Effect" as a verb exists -- "effect
		// change" -- but not with a pronoun or determiner after it, which is
		// why the object is matched too.
		Pattern: `\b(will|would|can|could|might|may|does|doesn't|didn't)\s+` +
			`effect\s+(the|a|an|my|your|his|her|our|their|you|me|us|them|him|it)\b`,
		Message:     "Should be \"$1 affect $2\"",
		Suggest:     "$1 affect $2",
		Category:    "Confused words",
		Explanation: explainAffectEffect,
	},

	// --- your / you're ---
	{
		// Four words, after a longer list was cut down word by word. Almost
		// everything that reads as an adjective after "you're" also reads as
		// something ownable after "your": your right (to refuse), your correct
		// address, your late payment, your next appointment, your invited
		// guests, your amazing work, your mistaken belief. Only these four
		// cannot follow a possessive at all.
		//
		// "Your welcome" is caught by its own rule among the error checks,
		// which predates this one.
		Pattern:     `\byour\s+(kidding|joking|probably|definitely)\b`,
		Message:     "Should be \"you're $1\"",
		Suggest:     "you're $1",
		Category:    "Confused words",
		Explanation: explainYourYoure,
	},
	{
		// The mirror: "you are" cannot own a noun.
		Pattern:     `\byou're\s+(own|car|house|turn|fault|money|wallet|keys|help|name|job)\b`,
		Message:     "Should be \"your $1\"",
		Suggest:     "your $1",
		Category:    "Confused words",
		Explanation: explainYourYoure,
	},

	// --- to / too ---
	{
		// The leading verb or intensifier is matched as well as the adjective
		// after it, because "to" before an adjective is fine on its own --
		// "spoke to many people", "went to great lengths".
		Pattern: `\b(is|are|was|were|be|been|it's|that's|feel|feels|felt|` +
			`seems|looks|way|far)\s+to\s+(much|many|late|early|big|small|` +
			`hard|easy|expensive|long|short|slow|fast|good|bad|old|young|` +
			`heavy|light|loud|quiet|soon|often)\b`,
		Message:     "Should be \"$1 too $2\"",
		Suggest:     "$1 too $2",
		Category:    "Confused words",
		Explanation: explainToToo,
	},
	{
		// "Me to." ending a sentence is always "me too". The antipattern is
		// the other reading: a "to" with something after it is the
		// preposition doing its job.
		Pattern:      `\b(me|him|her|us|them)\s+to\b`,
		Antipatterns: []string{`\b(me|him|her|us|them)\s+to\s+\S`},
		Message:      "Should be \"$1 too\"",
		Suggest:      "$1 too",
		Category:     "Confused words",
		Explanation:  explainToToo,
	},

	// --- their / they're ---
	{
		Pattern:     `\bthey're\s+(own|car|house|turn|fault|money|keys|wallet|names)\b`,
		Message:     "Should be \"their $1\"",
		Suggest:     "their $1",
		Category:    "Confused words",
		Explanation: explainTheirTheyre,
	},

	// --- whose / who's ---
	{
		Pattern: `\bwhose\s+(going|coming|been|gone|got|getting|ready|next|` +
			`there|here|right|wrong|that|this)\b`,
		Message:     "Should be \"who's $1\"",
		Suggest:     "who's $1",
		Category:    "Confused words",
		Explanation: explainWhoseWhos,
	},
	{
		Pattern:     `\bwho's\s+(turn|fault|idea|name|car|house|job|book|keys|phone)\b`,
		Message:     "Should be \"whose $1\"",
		Suggest:     "whose $1",
		Category:    "Confused words",
		Explanation: explainWhoseWhos,
	},

	// --- past participles after have/has/had ---
	// Each of these is wrong in every context: the participle is fixed by the
	// auxiliary, whatever the sentence is about.
	{
		Pattern:     `\b(have|has|had|having)\s+went\b`,
		Message:     "Should be \"$1 gone\"",
		Suggest:     "$1 gone",
		Category:    "Grammar",
		Explanation: explainParticiple,
	},
	{
		Pattern:     `\b(have|has|had|having)\s+came\b`,
		Message:     "Should be \"$1 come\"",
		Suggest:     "$1 come",
		Category:    "Grammar",
		Explanation: explainParticiple,
	},
	{
		Pattern:     `\b(have|has|had|having)\s+saw\b`,
		Message:     "Should be \"$1 seen\"",
		Suggest:     "$1 seen",
		Category:    "Grammar",
		Explanation: explainParticiple,
	},
	{
		Pattern:     `\b(have|has|had|having)\s+did\b`,
		Message:     "Should be \"$1 done\"",
		Suggest:     "$1 done",
		Category:    "Grammar",
		Explanation: explainParticiple,
	},
	{
		Pattern:     `\b(have|has|had|having)\s+wrote\b`,
		Message:     "Should be \"$1 written\"",
		Suggest:     "$1 written",
		Category:    "Grammar",
		Explanation: explainParticiple,
	},
	{
		Pattern:     `\b(have|has|had|having)\s+took\b`,
		Message:     "Should be \"$1 taken\"",
		Suggest:     "$1 taken",
		Category:    "Grammar",
		Explanation: explainParticiple,
	},
	{
		Pattern:     `\b(have|has|had|having)\s+gave\b`,
		Message:     "Should be \"$1 given\"",
		Suggest:     "$1 given",
		Category:    "Grammar",
		Explanation: explainParticiple,
	},
	{
		Pattern:     `\b(have|has|had|having)\s+broke\b`,
		Message:     "Should be \"$1 broken\"",
		Suggest:     "$1 broken",
		Category:    "Grammar",
		Explanation: explainParticiple,
	},
	{
		Pattern:     `\b(have|has|had|having)\s+spoke\b`,
		Message:     "Should be \"$1 spoken\"",
		Suggest:     "$1 spoken",
		Category:    "Grammar",
		Explanation: explainParticiple,
	},
	{
		Pattern:     `\b(have|has|had|having)\s+chose\b`,
		Message:     "Should be \"$1 chosen\"",
		Suggest:     "$1 chosen",
		Category:    "Grammar",
		Explanation: explainParticiple,
	},
	{
		Pattern:     `\b(have|has|had|having)\s+drank\b`,
		Message:     "Should be \"$1 drunk\"",
		Suggest:     "$1 drunk",
		Category:    "Grammar",
		Explanation: explainParticiple,
	},
	{
		// "Has lead to" is the commonest of these by a distance, because
		// "lead" the metal is spelled like the past tense sounds.
		Pattern:     `\b(have|has|had|having)\s+lead\s+to\b`,
		Message:     "Should be \"$1 led to\"",
		Suggest:     "$1 led to",
		Category:    "Grammar",
		Explanation: explainParticiple,
	},

	// --- phrases that are simply not the phrase ---
	{
		Pattern:  `\bapart\s+of\s+(the|a|an|my|your|our|their|this|that|it)\b`,
		Message:  "Should be \"a part of $1\"",
		Suggest:  "a part of $1",
		Category: "Confused words",
		Explanation: "\"Apart\" means separate, which is the opposite of " +
			"what belonging to something means.",
	},
	{
		Pattern:  `\bbare\s+with\s+(me|us)\b`,
		Message:  "Should be \"bear with $1\"",
		Suggest:  "bear with $1",
		Category: "Confused words",
		Explanation: "\"Bear\" here means to endure. \"Bare\" means uncovered, " +
			"which is a different request entirely.",
	},
	{
		Pattern:  `\bcould\s+care\s+less\b`,
		Message:  "Should be \"couldn't care less\"",
		Suggest:  "couldn't care less",
		Category: "Confused words",
		Explanation: "The phrase means you care so little that caring less " +
			"is impossible, which needs the negative.",
	},
	{
		Pattern:  `\bfor\s+all\s+intensive\s+purposes\b`,
		Message:  "Should be \"for all intents and purposes\"",
		Suggest:  "for all intents and purposes",
		Category: "Confused words",
		Explanation: "The phrase is \"intents and purposes\"; \"intensive\" " +
			"is a mishearing of it.",
	},
	{
		Pattern:  `\birregardless\b`,
		Message:  "Should be \"regardless\"",
		Suggest:  "regardless",
		Category: "Confused words",
		Explanation: "\"Regardless\" already means without regard; the " +
			"\"ir-\" prefix negates it a second time.",
	},
	{
		Pattern:  `\bsneak\s+peak\b`,
		Message:  "Should be \"sneak peek\"",
		Suggest:  "sneak peek",
		Category: "Confused words",
		Explanation: "A \"peek\" is a quick look. A \"peak\" is the top of " +
			"a mountain.",
	},
	{
		Pattern:  `\b(peak|peek)\s+(your|my|his|her|their|our)\s+interest\b`,
		Message:  "Should be \"pique $2 interest\"",
		Suggest:  "pique $2 interest",
		Category: "Confused words",
		Explanation: "\"Pique\" means to arouse or provoke. The other two " +
			"are a mountain top and a quick look.",
	},
	{
		Pattern:  `\bin\s+the\s+passed\b`,
		Message:  "Should be \"in the past\"",
		Suggest:  "in the past",
		Category: "Confused words",
		Explanation: "\"Past\" is the noun for time gone by. \"Passed\" is " +
			"the past tense of \"to pass\".",
	},
	{
		// Only at the end of a clause. "Principal" is also an adjective, so
		// the antipattern is the reading where a noun follows it: "in
		// principal cities", "on principal amounts".
		Pattern:      `\b(in|on)\s+principal\b`,
		Antipatterns: []string{`\b(in|on)\s+principal\s+\w`},
		Message:      "Should be \"$1 principle\"",
		Suggest:      "$1 principle",
		Category:     "Confused words",
		Explanation: "A \"principle\" is a rule or belief. A \"principal\" " +
			"is the head of something, or the main one.",
	},
	{
		Pattern:  `\b(some|good|bad|any|my|your|his|her|their|the)\s+advise\b`,
		Message:  "Should be \"$1 advice\"",
		Suggest:  "$1 advice",
		Category: "Confused words",
		Explanation: "\"Advice\" is the noun -- the thing given. \"Advise\" " +
			"is the verb -- the act of giving it.",
	},
	{
		Pattern:     `\b(to|will|would|can|could|please)\s+advice\b`,
		Message:     "Should be \"$1 advise\"",
		Suggest:     "$1 advise",
		Category:    "Confused words",
		Explanation: "\"Advise\" is the verb -- the act of giving advice.",
	},
}
