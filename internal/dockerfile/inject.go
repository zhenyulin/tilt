package dockerfile

import (
	"github.com/docker/distribution/reference"
	"github.com/windmilleng/tilt/internal/container"
)

func InjectImageDigest(df Dockerfile, selector container.RefSelector, ref reference.NamedTagged) (Dockerfile, bool, error) {
	ast, err := ParseAST(df)
	if err != nil {
		return "", false, err
	}

	modified, err := ast.InjectImageDigest(selector, ref)
	if err != nil {
		return "", false, err
	}

	if !modified {
		return df, false, nil
	}

	newDf, err := ast.Print()
	return newDf, true, err
}

// func InjectImageRegistry(df Dockerfile, rep model.RegistryReplacement) (Dockerfile, bool, error) {
// 	ast, err := ParseAST(df)
// 	if err != nil {
// 		return "", false, err
// 	}

// 	modified := false

// 	err = ast.traverseImageRefs(func(node *parser.Node, toReplace reference.Named) reference.Named {
// 		if reference.Domain(toReplace) == rep.Old {
// 			rest := reference.Path(toReplace)
// 			cleanRef := reference.TrimNamed(toReplace)

// 			new, err := reference.ParseNamed(fmt.Sprintf("%s/%s", rep.New, rest))
// 			if err != nil {
// 				fmt.Printf("%v\n", err)
// 				return nil
// 			}
// 			modified = true
// 			return new
// 		}
// 		return nil
// 	})

// 	df, err := ast.Print()
// 	if err != nil {
// 		return
// 	}

// 	return ast.Print(), modified, err
// }
