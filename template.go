package cedar

import (
    internalast "github.com/cedar-policy/cedar-go/internal/ast"
    "github.com/cedar-policy/cedar-go/internal/parser"
    "github.com/cedar-policy/cedar-go/types"
)

type Template Policy

type LinkedPolicy parser.LinkedPolicy

func LinkTemplate(template Template, templateID string, linkID string, slotEnv map[types.SlotID]types.EntityUID) LinkedPolicy {
    t := parser.Template(*template.ast)
    linkedPolicy := parser.NewLinkedPolicy(&t, templateID, linkID, slotEnv)

    return LinkedPolicy(linkedPolicy)
}

func (p LinkedPolicy) Render() (*Policy, error) {
    pl := parser.LinkedPolicy(p)

    policy, err := pl.Render()
    if err != nil {
        return nil, err
    }

    internalPolicy := internalast.Policy(policy)

    return newPolicy(&internalPolicy), nil
}
