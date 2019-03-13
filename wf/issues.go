package wf

import "github.com/lyraproj/issue/issue"

const (
	ConditionSyntaxError   = `WF_CONDITION_SYNTAX_ERROR`
	ConditionMissingRp     = `WF_CONDITION_MISSING_RP`
	ConditionInvalidName   = `WF_CONDITION_INVALID_NAME`
	ConditionUnexpectedEnd = `WF_CONDITION_UNEXPECTED_END`
	IllegalIterationStyle  = `WF_ILLEGAL_ITERATION_STYLE`
	IllegalOperation       = `WF_ILLEGAL_OPERATION`
	ActivityNoName         = `WF_ACTIVITY_NO_NAME`
	IteratorNotOneActivity = `WF_ITERATOR_NOT_ONE_ACTIVITY`
)

func init() {
	issue.Hard(ConditionSyntaxError, `syntax error in condition '%{text}' at position %{pos}`)
	issue.Hard(ConditionMissingRp, `expected right parenthesis in condition '%{text}' at position %{pos}`)
	issue.Hard(ConditionInvalidName, `invalid name '%{name}' in condition '%{text}' at position %{pos}`)
	issue.Hard(ConditionUnexpectedEnd, `unexpected end of condition '%{text}' at position %{pos}`)
	issue.Hard(IllegalIterationStyle, `no such iteration style '%{style}'`)
	issue.Hard(IllegalOperation, `no such operation '%{operation}'`)
	issue.Hard(ActivityNoName, `an activity must have a name`)
	issue.Hard(IteratorNotOneActivity, `an iterator must have exactly one activity`)
}
