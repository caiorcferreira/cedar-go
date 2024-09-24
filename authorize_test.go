package cedar_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go"
	"github.com/cedar-policy/cedar-go/internal/testutil"
)

//nolint:revive // due to table test function-length
func TestIsAuthorized(t *testing.T) {
	t.Parallel()
	cuzco := cedar.NewEntityUID("coder", "cuzco")
	dropTable := cedar.NewEntityUID("table", "drop")
	tests := []struct {
		Name                        string
		Policy                      string
		Entities                    cedar.Entities
		Principal, Action, Resource cedar.EntityUID
		Context                     cedar.Record
		Want                        cedar.Decision
		DiagErr                     int
		ParseErr                    bool
	}{
		{
			Name:      "simple-permit",
			Policy:    `permit(principal,action,resource);`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "simple-forbid",
			Policy:    `forbid(principal,action,resource);`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      false,
			DiagErr:   0,
		},
		{
			Name:      "no-permit",
			Policy:    `permit(principal,action,resource in asdf::"1234");`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      false,
			DiagErr:   0,
		},
		{
			Name:      "error-in-policy",
			Policy:    `permit(principal,action,resource) when { resource in "foo" };`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      false,
			DiagErr:   1,
		},
		{
			Name: "error-in-policy-continues",
			Policy: `permit(principal,action,resource) when { resource in "foo" };
			permit(principal,action,resource);
			`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      true,
			DiagErr:   1,
		},
		{
			Name:      "permit-requires-context-success",
			Policy:    `permit(principal,action,resource) when { context.x == 42 };`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.NewRecord(cedar.RecordMap{"x": cedar.Long(42)}),
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-requires-context-fail",
			Policy:    `permit(principal,action,resource) when { context.x == 42 };`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.NewRecord(cedar.RecordMap{"x": cedar.Long(43)}),
			Want:      false,
			DiagErr:   0,
		},
		{
			Name:   "permit-requires-entities-success",
			Policy: `permit(principal,action,resource) when { principal.x == 42 };`,
			Entities: cedar.Entities{
				cuzco: &cedar.Entity{
					UID:        cuzco,
					Attributes: cedar.NewRecord(cedar.RecordMap{"x": cedar.Long(42)}),
				},
			},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:   "permit-requires-entities-fail",
			Policy: `permit(principal,action,resource) when { principal.x == 42 };`,
			Entities: cedar.Entities{
				cuzco: &cedar.Entity{
					UID:        cuzco,
					Attributes: cedar.NewRecord(cedar.RecordMap{"x": cedar.Long(43)}),
				},
			},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      false,
			DiagErr:   0,
		},
		{
			Name:   "permit-requires-entities-parent-success",
			Policy: `permit(principal,action,resource) when { principal in parent::"bob" };`,
			Entities: cedar.Entities{
				cuzco: &cedar.Entity{
					UID:     cuzco,
					Parents: *cedar.NewEntityUIDSetFromSlice([]cedar.EntityUID{cedar.NewEntityUID("parent", "bob")}),
				},
			},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-principal-equals",
			Policy:    `permit(principal == coder::"cuzco",action,resource);`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:   "permit-principal-in",
			Policy: `permit(principal in team::"osiris",action,resource);`,
			Entities: cedar.Entities{
				cuzco: &cedar.Entity{
					UID:     cuzco,
					Parents: *cedar.NewEntityUIDSetFromSlice([]cedar.EntityUID{cedar.NewEntityUID("team", "osiris")}),
				},
			},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-action-equals",
			Policy:    `permit(principal,action == table::"drop",resource);`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:   "permit-action-in",
			Policy: `permit(principal,action in scary::"stuff",resource);`,
			Entities: cedar.Entities{
				dropTable: &cedar.Entity{
					UID:     dropTable,
					Parents: *cedar.NewEntityUIDSetFromSlice([]cedar.EntityUID{cedar.NewEntityUID("scary", "stuff")}),
				},
			},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:   "permit-action-in-set",
			Policy: `permit(principal,action in [scary::"stuff"],resource);`,
			Entities: cedar.Entities{
				dropTable: &cedar.Entity{
					UID:     dropTable,
					Parents: *cedar.NewEntityUIDSetFromSlice([]cedar.EntityUID{cedar.NewEntityUID("scary", "stuff")}),
				},
			},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-resource-equals",
			Policy:    `permit(principal,action,resource == table::"whatever");`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-unless",
			Policy:    `permit(principal,action,resource) unless { false };`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-if",
			Policy:    `permit(principal,action,resource) when { (if true then true else true) };`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-or",
			Policy:    `permit(principal,action,resource) when { (true || false) };`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-and",
			Policy:    `permit(principal,action,resource) when { (true && true) };`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-relations",
			Policy:    `permit(principal,action,resource) when { (1<2) && (1<=1) && (2>1) && (1>=1) && (1!=2) && (1==1)};`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-relations-in",
			Policy:    `permit(principal,action,resource) when { principal in principal };`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:   "permit-when-relations-has",
			Policy: `permit(principal,action,resource) when { principal has name };`,
			Entities: cedar.Entities{
				cuzco: &cedar.Entity{
					UID:        cuzco,
					Attributes: cedar.NewRecord(cedar.RecordMap{"name": cedar.String("bob")}),
				},
			},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-add-sub",
			Policy:    `permit(principal,action,resource) when { 40+3-1==42 };`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-mul",
			Policy:    `permit(principal,action,resource) when { 6*7==42 };`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-negate",
			Policy:    `permit(principal,action,resource) when { -42==-42 };`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-not",
			Policy:    `permit(principal,action,resource) when { !(1+1==42) };`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-set",
			Policy:    `permit(principal,action,resource) when { [1,2,3].contains(2) };`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-record",
			Policy:    `permit(principal,action,resource) when { {name:"bob"} has name };`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-action",
			Policy:    `permit(principal,action,resource) when { action in action };`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-set-contains-ok",
			Policy:    `permit(principal,action,resource) when { [1,2,3].contains(2) };`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-set-contains-error",
			Policy:    `permit(principal,action,resource) when { [1,2,3].contains(2,3) };`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      false,
			DiagErr:   0,
			ParseErr:  true,
		},
		{
			Name:      "permit-when-set-containsAll-ok",
			Policy:    `permit(principal,action,resource) when { [1,2,3].containsAll([2,3]) };`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-set-containsAll-error",
			Policy:    `permit(principal,action,resource) when { [1,2,3].containsAll(2,3) };`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      false,
			DiagErr:   0,
			ParseErr:  true,
		},
		{
			Name:      "permit-when-set-containsAny-ok",
			Policy:    `permit(principal,action,resource) when { [1,2,3].containsAny([2,5]) };`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-set-containsAny-error",
			Policy:    `permit(principal,action,resource) when { [1,2,3].containsAny(2,5) };`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      false,
			DiagErr:   0,
			ParseErr:  true,
		},
		{
			Name:      "permit-when-record-attr",
			Policy:    `permit(principal,action,resource) when { {name:"bob"}["name"] == "bob" };`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-unknown-method",
			Policy:    `permit(principal,action,resource) when { [1,2,3].shuffle() };`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      false,
			DiagErr:   0,
			ParseErr:  true,
		},
		{
			Name:      "permit-when-like",
			Policy:    `permit(principal,action,resource) when { "bananas" like "*nan*" };`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-unknown-ext-fun",
			Policy:    `permit(principal,action,resource) when { fooBar("10") };`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      false,
			DiagErr:   0,
			ParseErr:  true,
		},
		{
			Name: "permit-when-decimal",
			Policy: `permit(principal,action,resource) when {
				decimal("10.0").lessThan(decimal("11.0")) &&
				decimal("10.0").lessThanOrEqual(decimal("11.0")) &&
				decimal("10.0").greaterThan(decimal("9.0")) &&
				decimal("10.0").greaterThanOrEqual(decimal("9.0")) };`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-decimal-fun-wrong-arity",
			Policy:    `permit(principal,action,resource) when { decimal(1, 2) };`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      false,
			DiagErr:   1,
		},
		{
			Name: "permit-when-datetime",
			Policy: `permit(principal,action,resource) when {
				datetime("1970-01-01T09:08:07Z") < (datetime("1970-02-01")) &&
				datetime("1970-01-01T09:08:07Z") <= (datetime("1970-02-01")) &&
				datetime("1970-01-01T09:08:07Z") > (datetime("1970-01-01")) &&
				datetime("1970-01-01T09:08:07Z") >= (datetime("1970-01-01")) &&
        datetime("1970-01-01T09:08:07Z").toDate() == datetime("1970-01-01")};`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-datetime-fun-wrong-arity",
			Policy:    `permit(principal,action,resource) when { datetime("1970-01-01", "UTC") };`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      false,
			DiagErr:   1,
		},
		{
			Name: "permit-when-duration",
			Policy: `permit(principal,action,resource) when {
				duration("9h8m") < (duration("10h")) &&
				duration("9h8m") <= (duration("10h")) &&
				duration("9h8m") > (duration("7h")) &&
				duration("9h8m") >= (duration("7h")) &&
				duration("1ms").toMilliseconds() == 1 &&
				duration("1s").toSeconds() == 1 &&
				duration("1m").toMinutes() == 1 &&
				duration("1h").toHours() == 1 &&
				duration("1d").toDays() == 1 &&
        datetime("1970-01-01").toTime() == duration("0ms") &&
        datetime("1970-01-01").offset(duration("1ms")).toTime() == duration("1ms") &&
        datetime("1970-01-01T00:00:00.001Z").durationSince(datetime("1970-01-01")) == duration("1ms")};`,

			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-duration-fun-wrong-arity",
			Policy:    `permit(principal,action,resource) when { duration("1h", "huh?") };`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      false,
			DiagErr:   1,
		},
		{
			Name: "permit-when-ip",
			Policy: `permit(principal,action,resource) when {
				ip("1.2.3.4").isIpv4() &&
				ip("a:b:c:d::/16").isIpv6() &&
				ip("::1").isLoopback() &&
				ip("224.1.2.3").isMulticast() &&
				ip("127.0.0.1").isInRange(ip("127.0.0.0/16"))};`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-ip-fun-wrong-arity",
			Policy:    `permit(principal,action,resource) when { ip() };`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      false,
			DiagErr:   1,
		},
		{
			Name:      "permit-when-isIpv4-wrong-arity",
			Policy:    `permit(principal,action,resource) when { ip("1.2.3.4").isIpv4(true) };`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      false,
			DiagErr:   1,
		},
		{
			Name:      "permit-when-isIpv6-wrong-arity",
			Policy:    `permit(principal,action,resource) when { ip("1.2.3.4").isIpv6(true) };`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      false,
			DiagErr:   1,
		},
		{
			Name:      "permit-when-isLoopback-wrong-arity",
			Policy:    `permit(principal,action,resource) when { ip("1.2.3.4").isLoopback(true) };`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      false,
			DiagErr:   1,
		},
		{
			Name:      "permit-when-isMulticast-wrong-arity",
			Policy:    `permit(principal,action,resource) when { ip("1.2.3.4").isMulticast(true) };`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      false,
			DiagErr:   1,
		},
		{
			Name:      "permit-when-isInRange-wrong-arity",
			Policy:    `permit(principal,action,resource) when { ip("1.2.3.4").isInRange() };`,
			Entities:  cedar.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      false,
			DiagErr:   1,
		},
		{
			Name:     "negative-unary-op",
			Policy:   `permit(principal,action,resource) when { -context.value > 0 };`,
			Entities: cedar.Entities{},
			Context:  cedar.NewRecord(cedar.RecordMap{"value": cedar.Long(-42)}),
			Want:     true,
			DiagErr:  0,
		},
		{
			Name:      "principal-is",
			Policy:    `permit(principal is Actor,action,resource);`,
			Entities:  cedar.Entities{},
			Principal: cedar.NewEntityUID("Actor", "cuzco"),
			Action:    cedar.NewEntityUID("Action", "drop"),
			Resource:  cedar.NewEntityUID("Resource", "table"),
			Context:   cedar.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "principal-is-in",
			Policy:    `permit(principal is Actor in Actor::"cuzco",action,resource);`,
			Entities:  cedar.Entities{},
			Principal: cedar.NewEntityUID("Actor", "cuzco"),
			Action:    cedar.NewEntityUID("Action", "drop"),
			Resource:  cedar.NewEntityUID("Resource", "table"),
			Context:   cedar.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "resource-is",
			Policy:    `permit(principal,action,resource is Resource);`,
			Entities:  cedar.Entities{},
			Principal: cedar.NewEntityUID("Actor", "cuzco"),
			Action:    cedar.NewEntityUID("Action", "drop"),
			Resource:  cedar.NewEntityUID("Resource", "table"),
			Context:   cedar.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "resource-is-in",
			Policy:    `permit(principal,action,resource is Resource in Resource::"table");`,
			Entities:  cedar.Entities{},
			Principal: cedar.NewEntityUID("Actor", "cuzco"),
			Action:    cedar.NewEntityUID("Action", "drop"),
			Resource:  cedar.NewEntityUID("Resource", "table"),
			Context:   cedar.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "when-is",
			Policy:    `permit(principal,action,resource) when { resource is Resource };`,
			Entities:  cedar.Entities{},
			Principal: cedar.NewEntityUID("Actor", "cuzco"),
			Action:    cedar.NewEntityUID("Action", "drop"),
			Resource:  cedar.NewEntityUID("Resource", "table"),
			Context:   cedar.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "when-is-in",
			Policy:    `permit(principal,action,resource) when { resource is Resource in Resource::"table" };`,
			Entities:  cedar.Entities{},
			Principal: cedar.NewEntityUID("Actor", "cuzco"),
			Action:    cedar.NewEntityUID("Action", "drop"),
			Resource:  cedar.NewEntityUID("Resource", "table"),
			Context:   cedar.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:   "when-is-in",
			Policy: `permit(principal,action,resource) when { resource is Resource in Parent::"id" };`,
			Entities: cedar.Entities{
				cedar.NewEntityUID("Resource", "table"): &cedar.Entity{
					UID:     cedar.NewEntityUID("Resource", "table"),
					Parents: *cedar.NewEntityUIDSetFromSlice([]cedar.EntityUID{cedar.NewEntityUID("Parent", "id")}),
				},
			},
			Principal: cedar.NewEntityUID("Actor", "cuzco"),
			Action:    cedar.NewEntityUID("Action", "drop"),
			Resource:  cedar.NewEntityUID("Resource", "table"),
			Context:   cedar.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:   "rfc-57", // https://github.com/cedar-policy/rfcs/blob/main/text/0057-general-multiplication.md
			Policy: `permit(principal, action, resource) when { context.foo * principal.bar >= 100 };`,
			Entities: cedar.Entities{
				cedar.NewEntityUID("Principal", "1"): &cedar.Entity{
					UID:        cedar.NewEntityUID("Principal", "1"),
					Attributes: cedar.NewRecord(cedar.RecordMap{"bar": cedar.Long(42)}),
				},
			},
			Principal: cedar.NewEntityUID("Principal", "1"),
			Action:    cedar.NewEntityUID("Action", "action"),
			Resource:  cedar.NewEntityUID("Resource", "resource"),
			Context:   cedar.NewRecord(cedar.RecordMap{"foo": cedar.Long(43)}),
			Want:      true,
			DiagErr:   0,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			ps, err := cedar.NewPolicySetFromBytes("policy.cedar", []byte(tt.Policy))
			testutil.Equals(t, err != nil, tt.ParseErr)
			ok, diag := ps.IsAuthorized(tt.Entities, cedar.Request{
				Principal: tt.Principal,
				Action:    tt.Action,
				Resource:  tt.Resource,
				Context:   tt.Context,
			})
			testutil.Equals(t, len(diag.Errors), tt.DiagErr)
			testutil.Equals(t, ok, tt.Want)
		})
	}
}
