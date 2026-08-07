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
		Flags:       []string{"this is better then that"},
		Leaves:      []string{"better than that"},
		Message:     "Should be \"$1 than\"",
		Suggest:     "$1 than",
		Category:    "Confused words",
		Explanation: explainThenThan,
	},
	{
		// The mirror. A conjunction cannot be followed by a comparison.
		Pattern:     `\b([aA]nd|[bB]ut|[sS]o|[bB]ack|[sS]ince|[rR]ight|[oO]nly|[wW]ay)\s+than\b`,
		Flags:       []string{"we synced and than sent it"},
		Leaves:      []string{"we synced and then sent it"},
		Message:     "Should be \"$1 then\"",
		Suggest:     "$1 then",
		Category:    "Confused words",
		Explanation: explainThenThan,
	},

	// --- lose / loose ---
	{
		// "Loosing" is not a word anyone means: the verb "to loose" exists
		// but its participle is vanishingly rare beside the typo.
		Pattern:     `\b[lL]oosing\b`,
		Flags:       []string{"we are loosing peers"},
		Message:     "Should be \"losing\"",
		Suggest:     "losing",
		Category:    "Confused words",
		Explanation: explainLoseLoose,
	},
	{
		// An adjective cannot take an object, so anything that looks like one
		// after "loose" means the verb was meant.
		Pattern: `\b[lL]oose\s+(the|a|an|my|your|his|her|our|their|it|them|` +
			`money|weight|control|track|interest|everything|access)\b`,
		Flags:       []string{"do not loose the keys"},
		Leaves:      []string{"the screw is loose here"},
		Message:     "Should be \"lose $1\"",
		Suggest:     "lose $1",
		Category:    "Confused words",
		Explanation: explainLoseLoose,
	},
	{
		// A modal or "to" forces the verb.
		Pattern:     `\b([wW]ill|[wW]ould|[cC]an|[cC]ould|[mM]ight|[mM]ay|[tT]o|[dD]on't|[dD]idn't|[dD]oesn't)\s+loose\b`,
		Flags:       []string{"you will loose it"},
		Message:     "Should be \"$1 lose\"",
		Suggest:     "$1 lose",
		Category:    "Confused words",
		Explanation: explainLoseLoose,
	},

	// --- affect / effect ---
	{
		// A determiner forces a noun.
		//
		// The indefinite article is separated out below, because correcting
		// the word changes which article the sentence needs: "a affect"
		// becomes "an effect", not "a effect". A rule that carries the
		// determiner through unchanged hands back a phrase that is still
		// wrong, and this one did.
		Pattern:     `\b([tT]he|[tT]his|[tT]hat|[aA]ny|[nN]o|[sS]ide|[nN]et|[lL]ittle|[gG]reat|[wW]hose)\s+affect\b`,
		Flags:       []string{"it had no affect"},
		Leaves:      []string{"it had no effect"},
		Message:     "Should be \"$1 effect\"",
		Suggest:     "$1 effect",
		Category:    "Confused words",
		Explanation: explainAffectEffect,
	},
	{
		// Both "a affect" and "an affect" become "an effect": the vowel is
		// what decides, and after the correction there is one.
		Pattern:     `\b([Aa])n?\s+affect\b`,
		Flags:       []string{"it had a affect", "it had an affect"},
		Message:     "Should be \"$1n effect\"",
		Suggest:     "$1n effect",
		Category:    "Confused words",
		Explanation: explainAffectEffect,
	},
	{
		Pattern:     `\b([tT]he|[tT]hese|[tT]hose|[iI]ts|[sS]ide|[aA]ny|[nN]o)\s+affects\b`,
		Flags:       []string{"the affects were clear"},
		Message:     "Should be \"$1 effects\"",
		Suggest:     "$1 effects",
		Category:    "Confused words",
		Explanation: explainAffectEffect,
	},
	{
		// A modal forces the verb. "Effect" as a verb exists -- "effect
		// change" -- but not with a pronoun or determiner after it, which is
		// why the object is matched too.
		Pattern: `\b([wW]ill|[wW]ould|[cC]an|[cC]ould|[mM]ight|[mM]ay|[dD]oes|[dD]oesn't|[dD]idn't)\s+` +
			`effect\s+(the|a|an|my|your|his|her|our|their|you|me|us|them|him|it)\b`,
		Flags:       []string{"it will effect the outcome"},
		Leaves:      []string{"it will effect change across the network"},
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
		Pattern:     `\b([yY])our\s+(kidding|joking|probably|definitely)\b`,
		Flags:       []string{"your kidding me", "Your joking"},
		Leaves:      []string{"it is your right to refuse"},
		Message:     "Should be \"$1ou're $2\"",
		Suggest:     "$1ou're $2",
		Category:    "Confused words",
		Explanation: explainYourYoure,
	},
	{
		// The mirror: "you are" cannot own a noun.
		Pattern:     `\b[yY]ou're\s+(own|car|house|turn|fault|money|wallet|keys|help|name|job)\b`,
		Flags:       []string{"that is you're own fault"},
		Message:     "Should be \"your $1\"",
		Suggest:     "your $1",
		Category:    "Confused words",
		Explanation: explainYourYoure,
	},

	{
		// "thanks for you message". The dropped r is a typing slip rather
		// than a confusion, which is why neither rule above catches it:
		// "you" is a perfectly good word and this is a perfectly good
		// sentence shape.
		//
		// What makes it decidable is the noun, not the preposition. "for
		// you" is complete on its own, so anything can follow it -- "I did
		// it for you yesterday" is correct, and a rule keyed on "preposition
		// + you + any noun" would flag it. The nouns below are the ones that
		// cannot stand bare after "you": there is no reading of "for you
		// message" or "with you permission" that works.
		Pattern: `\b([Ff]or|[Ww]ith|[Aa]bout|[Ii]n|[Oo]n)\s+you\s+(message|` +
			`email|reply|response|time|help|support|feedback|patience|` +
			`understanding|interest|advice|input|permission|opinion|` +
			`behalf|order|payment|invoice|account|address|question|` +
			`questions|comments|thoughts|notes|post|work|effort|kindness)\b`,
		Flags: []string{
			"thanks for you message",
			"Hi Sam, thanks for you message",
			"with you permission we will proceed",
		},
		Leaves: []string{
			"I left it for you yesterday",
			"this is for you and nobody else",
			"we are counting on you today",
		},
		Message:     "Should be \"$1 your $2\"",
		Suggest:     "$1 your $2",
		Category:    "Confused words",
		Explanation: "\"You\" is the pronoun and \"your\" is the possessive. The noun after it is being owned, so this needs \"your\".",
	},

	// --- to / too ---
	{
		// The leading verb or intensifier is matched as well as the adjective
		// after it, because "to" before an adjective is fine on its own --
		// "spoke to many people", "went to great lengths".
		//
		// The adjectives are listed rather than matched loosely, and the
		// list is the whole rule. "High" was missing until an example asked
		// for it, which is about as common as this mistake gets. Words that
		// are also verbs stay out: "the plan is to close the door" and "the
		// aim is to narrow the gap" are correct writing that "close" or
		// "narrow" in this list would flag.
		Pattern: `\b(is|are|was|were|be|been|it's|that's|feel|feels|felt|` +
			`seems|looks|way|far)\s+to\s+(much|many|late|early|big|small|` +
			`hard|easy|expensive|long|short|slow|fast|good|bad|old|young|` +
			`heavy|light|loud|quiet|soon|often|high|low|hot|cold|busy|` +
			`tired|strong|weak|wide|deep|thin|thick|dark|bright|complex|` +
			`simple|risky|similar|different)\b`,
		Flags:       []string{"the fee is to high"},
		Leaves:      []string{"i spoke to many people about it"},
		Message:     "Should be \"$1 too $2\"",
		Suggest:     "$1 too $2",
		Category:    "Confused words",
		Explanation: explainToToo,
	},
	{
		// "Me to." ending a sentence is always "me too". The antipattern is
		// the other reading: a "to" with something after it is the
		// preposition doing its job.
		Pattern:      `\b([mM]e|[hH]im|[hH]er|[uU]s|[tT]hem)\s+to\b`,
		Flags:        []string{"send it to me to."},
		Leaves:       []string{"send it to me to review"},
		Antipatterns: []string{`\b([mM]e|[hH]im|[hH]er|[uU]s|[tT]hem)\s+to\s+\S`},
		Message:      "Should be \"$1 too\"",
		Suggest:      "$1 too",
		Category:     "Confused words",
		Explanation:  explainToToo,
	},

	// --- their / they're ---
	{
		Pattern:     `\b[tT]hey're\s+(own|car|house|turn|fault|money|keys|wallet|names)\b`,
		Flags:       []string{"they're own fault"},
		Message:     "Should be \"their $1\"",
		Suggest:     "their $1",
		Category:    "Confused words",
		Explanation: explainTheirTheyre,
	},

	// --- whose / who's ---
	{
		Pattern: `\b[wW]hose\s+(going|coming|been|gone|got|getting|ready|next|` +
			`there|here|right|wrong|that|this)\b`,
		Flags:       []string{"whose going to test it"},
		Leaves:      []string{"whose keys are these"},
		Message:     "Should be \"who's $1\"",
		Suggest:     "who's $1",
		Category:    "Confused words",
		Explanation: explainWhoseWhos,
	},
	{
		Pattern:     `\b[wW]ho's\s+(turn|fault|idea|name|car|house|job|book|keys|phone)\b`,
		Flags:       []string{"who's turn is it"},
		Leaves:      []string{"who's going to review it"},
		Message:     "Should be \"whose $1\"",
		Suggest:     "whose $1",
		Category:    "Confused words",
		Explanation: explainWhoseWhos,
	},

	// --- past participles after have/has/had ---
	// Each of these is wrong in every context: the participle is fixed by the
	// auxiliary, whatever the sentence is about.
	{
		Pattern:     `\b([hH]ave|[hH]as|[hH]ad|[hH]aving)\s+went\b`,
		Flags:       []string{"they have went home"},
		Leaves:      []string{"they have gone home"},
		Message:     "Should be \"$1 gone\"",
		Suggest:     "$1 gone",
		Category:    "Grammar",
		Explanation: explainParticiple,
	},
	{
		Pattern:     `\b([hH]ave|[hH]as|[hH]ad|[hH]aving)\s+came\b`,
		Flags:       []string{"they have came back"},
		Message:     "Should be \"$1 come\"",
		Suggest:     "$1 come",
		Category:    "Grammar",
		Explanation: explainParticiple,
	},
	{
		Pattern:     `\b([hH]ave|[hH]as|[hH]ad|[hH]aving)\s+saw\b`,
		Flags:       []string{"i have saw it"},
		Message:     "Should be \"$1 seen\"",
		Suggest:     "$1 seen",
		Category:    "Grammar",
		Explanation: explainParticiple,
	},
	{
		Pattern:     `\b([hH]ave|[hH]as|[hH]ad|[hH]aving)\s+did\b`,
		Flags:       []string{"i have did that"},
		Message:     "Should be \"$1 done\"",
		Suggest:     "$1 done",
		Category:    "Grammar",
		Explanation: explainParticiple,
	},
	{
		Pattern:     `\b([hH]ave|[hH]as|[hH]ad|[hH]aving)\s+wrote\b`,
		Flags:       []string{"i have wrote the post"},
		Message:     "Should be \"$1 written\"",
		Suggest:     "$1 written",
		Category:    "Grammar",
		Explanation: explainParticiple,
	},
	{
		Pattern:     `\b([hH]ave|[hH]as|[hH]ad|[hH]aving)\s+took\b`,
		Flags:       []string{"i have took it"},
		Message:     "Should be \"$1 taken\"",
		Suggest:     "$1 taken",
		Category:    "Grammar",
		Explanation: explainParticiple,
	},
	{
		Pattern:     `\b([hH]ave|[hH]as|[hH]ad|[hH]aving)\s+gave\b`,
		Flags:       []string{"i have gave it"},
		Message:     "Should be \"$1 given\"",
		Suggest:     "$1 given",
		Category:    "Grammar",
		Explanation: explainParticiple,
	},
	{
		Pattern:     `\b([hH]ave|[hH]as|[hH]ad|[hH]aving)\s+broke\b`,
		Flags:       []string{"i have broke it"},
		Message:     "Should be \"$1 broken\"",
		Suggest:     "$1 broken",
		Category:    "Grammar",
		Explanation: explainParticiple,
	},
	{
		Pattern:     `\b([hH]ave|[hH]as|[hH]ad|[hH]aving)\s+spoke\b`,
		Flags:       []string{"i have spoke to them"},
		Message:     "Should be \"$1 spoken\"",
		Suggest:     "$1 spoken",
		Category:    "Grammar",
		Explanation: explainParticiple,
	},
	{
		Pattern:     `\b([hH]ave|[hH]as|[hH]ad|[hH]aving)\s+chose\b`,
		Flags:       []string{"i have chose one"},
		Message:     "Should be \"$1 chosen\"",
		Suggest:     "$1 chosen",
		Category:    "Grammar",
		Explanation: explainParticiple,
	},
	{
		Pattern:     `\b([hH]ave|[hH]as|[hH]ad|[hH]aving)\s+drank\b`,
		Flags:       []string{"i have drank it"},
		Message:     "Should be \"$1 drunk\"",
		Suggest:     "$1 drunk",
		Category:    "Grammar",
		Explanation: explainParticiple,
	},
	{
		// "Has lead to" is the commonest of these by a distance, because
		// "lead" the metal is spelled like the past tense sounds.
		Pattern:     `\b([hH]ave|[hH]as|[hH]ad|[hH]aving)\s+lead\s+to\b`,
		Flags:       []string{"that has lead to problems"},
		Message:     "Should be \"$1 led to\"",
		Suggest:     "$1 led to",
		Category:    "Grammar",
		Explanation: explainParticiple,
	},

	// --- phrases that are simply not the phrase ---
	{
		Pattern:  `\b[aA]part\s+of\s+(the|a|an|my|your|our|their|this|that|it)\b`,
		Flags:    []string{"it was apart of the plan"},
		Leaves:   []string{"apart from that it works"},
		Message:  "Should be \"a part of $1\"",
		Suggest:  "a part of $1",
		Category: "Confused words",
		Explanation: "\"Apart\" means separate, which is the opposite of " +
			"what belonging to something means.",
	},
	{
		Pattern:  `\b[bB]are\s+with\s+(me|us)\b`,
		Flags:    []string{"bare with me a moment"},
		Leaves:   []string{"bear with me a moment"},
		Message:  "Should be \"bear with $1\"",
		Suggest:  "bear with $1",
		Category: "Confused words",
		Explanation: "\"Bear\" here means to endure. \"Bare\" means uncovered, " +
			"which is a different request entirely.",
	},
	{
		Pattern:  `\b[cC]ould\s+care\s+less\b`,
		Flags:    []string{"i could care less"},
		Message:  "Should be \"couldn't care less\"",
		Suggest:  "couldn't care less",
		Category: "Confused words",
		Explanation: "The phrase means you care so little that caring less " +
			"is impossible, which needs the negative.",
	},
	{
		Pattern:  `\b[fF]or\s+all\s+intensive\s+purposes\b`,
		Flags:    []string{"for all intensive purposes"},
		Message:  "Should be \"for all intents and purposes\"",
		Suggest:  "for all intents and purposes",
		Category: "Confused words",
		Explanation: "The phrase is \"intents and purposes\"; \"intensive\" " +
			"is a mishearing of it.",
	},
	{
		Pattern:  `\b[iI]rregardless\b`,
		Flags:    []string{"irregardless of the cost"},
		Message:  "Should be \"regardless\"",
		Suggest:  "regardless",
		Category: "Confused words",
		Explanation: "\"Regardless\" already means without regard; the " +
			"\"ir-\" prefix negates it a second time.",
	},
	{
		Pattern:  `\b[sS]neak\s+peak\b`,
		Flags:    []string{"a sneak peak at it"},
		Message:  "Should be \"sneak peek\"",
		Suggest:  "sneak peek",
		Category: "Confused words",
		Explanation: "A \"peek\" is a quick look. A \"peak\" is the top of " +
			"a mountain.",
	},
	{
		Pattern:  `\b([pP]eak|[pP]eek)\s+(your|my|his|her|their|our)\s+interest\b`,
		Flags:    []string{"peak your interest"},
		Message:  "Should be \"pique $2 interest\"",
		Suggest:  "pique $2 interest",
		Category: "Confused words",
		Explanation: "\"Pique\" means to arouse or provoke. The other two " +
			"are a mountain top and a quick look.",
	},
	{
		Pattern:  `\b[iI]n\s+the\s+passed\b`,
		Flags:    []string{"in the passed we did"},
		Leaves:   []string{"in the past we did"},
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
		Pattern:      `\b([iI]n|[oO]n)\s+principal\b`,
		Flags:        []string{"i agree in principal."},
		Leaves:       []string{"the bank operates in principal cities"},
		Antipatterns: []string{`\b([iI]n|[oO]n)\s+principal\s+\w`},
		Message:      "Should be \"$1 principle\"",
		Suggest:      "$1 principle",
		Category:     "Confused words",
		Explanation: "A \"principle\" is a rule or belief. A \"principal\" " +
			"is the head of something, or the main one.",
	},
	{
		Pattern:  `\b([sS]ome|[gG]ood|[bB]ad|[aA]ny|[mM]y|[yY]our|[hH]is|[hH]er|[tT]heir|[tT]he)\s+advise\b`,
		Flags:    []string{"thanks for the advise"},
		Leaves:   []string{"thanks for the advice"},
		Message:  "Should be \"$1 advice\"",
		Suggest:  "$1 advice",
		Category: "Confused words",
		Explanation: "\"Advice\" is the noun -- the thing given. \"Advise\" " +
			"is the verb -- the act of giving it.",
	},
	{
		Pattern:     `\b([tT]o|[wW]ill|[wW]ould|[cC]an|[cC]ould|[pP]lease)\s+advice\b`,
		Flags:       []string{"please advice us"},
		Leaves:      []string{"please advise us"},
		Message:     "Should be \"$1 advise\"",
		Suggest:     "$1 advise",
		Category:    "Confused words",
		Explanation: "\"Advise\" is the verb -- the act of giving advice.",
	},
}
