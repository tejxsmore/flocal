//#region src/lib/types.ts
var FORMAT_OPTIONS = [
	{
		label: "Word",
		value: "word",
		icon: "📝"
	},
	{
		label: "Quote",
		value: "quote",
		icon: "💭"
	},
	{
		label: "Debate",
		value: "debate",
		icon: "⚖️"
	},
	{
		label: "Situation",
		value: "situation",
		icon: "🎭"
	},
	{
		label: "Story Starter",
		value: "story_starter",
		icon: "📖"
	},
	{
		label: "Image",
		value: "image",
		icon: "📸"
	}
];
var PREP_TIME_OPTIONS = [
	{
		label: "No prep",
		value: 0
	},
	{
		label: "5 min",
		value: 300
	},
	{
		label: "10 min",
		value: 600
	},
	{
		label: "15 min",
		value: 900
	}
];
var SORT_OPTIONS = [
	{
		label: "Newest first",
		value: "newest"
	},
	{
		label: "Oldest first",
		value: "oldest"
	},
	{
		label: "Highest score",
		value: "score_desc"
	},
	{
		label: "Lowest score",
		value: "score_asc"
	}
];
var STATUS_OPTIONS = [
	{
		label: "Completed",
		value: "completed"
	},
	{
		label: "Processing",
		value: "processing"
	},
	{
		label: "Pending",
		value: "pending"
	},
	{
		label: "Failed",
		value: "failed"
	}
];
//#endregion
export { STATUS_OPTIONS as i, PREP_TIME_OPTIONS as n, SORT_OPTIONS as r, FORMAT_OPTIONS as t };
