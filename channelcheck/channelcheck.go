// Package addcheck defines an Analyzer that reports time package expressions that
// can be simplified
package channelcheck

/*
	Strategy: Identify proper SendStmt's early. If we don't see it within a Select, then it must be used improper.
	 Or, we simply MISS it, making it a false positive which is fine. We'd rather fail open and get the user to
	 use nolints than miss a potential bug altogether.

	Steps for blocking channel send:
	- Find a SelectStmt
	- Check if Comm Clause:
		- See if it has a 'ast.SendStmt'.
		- If it does, then check 'default' or 'ticker' types there too.
	- If we find SelectStmt otherwise, it must be invalid.

	TODOs:
	- Flags for EACH linter to enable/disable it.
	- Timer/timeout support as alternative default cases. Or, just something that is NOT a channel send maybe?
	- Integrate with the golangci-lint system.

	Inspired by John's PoC implementation: https://github.com/johnsaigle/channelcheck

	Example AST
	==========================

     0  *ast.SelectStmt {
     1  .  Select: -
     2  .  Body: *ast.BlockStmt {
     3  .  .  Lbrace: -
     4  .  .  List: []ast.Stmt (len = 2) {
     5  .  .  .  0: *ast.CommClause {
     6  .  .  .  .  Case: -
     7  .  .  .  .  Comm: *ast.SendStmt {
     8  .  .  .  .  .  Chan: *ast.Ident {
     9  .  .  .  .  .  .  NamePos: -
    10  .  .  .  .  .  .  Name: "c"
    11  .  .  .  .  .  .  Obj: *ast.Object {
    12  .  .  .  .  .  .  .  Kind: var
    13  .  .  .  .  .  .  .  Name: "c"
    14  .  .  .  .  .  .  .  Decl: *ast.Field {
    15  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
    16  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
    17  .  .  .  .  .  .  .  .  .  .  NamePos: -
    18  .  .  .  .  .  .  .  .  .  .  Name: "c"
    19  .  .  .  .  .  .  .  .  .  .  Obj: *(obj @ 11)
    20  .  .  .  .  .  .  .  .  .  }
    21  .  .  .  .  .  .  .  .  }
    22  .  .  .  .  .  .  .  .  Type: *ast.ChanType {
    23  .  .  .  .  .  .  .  .  .  Begin: -
    24  .  .  .  .  .  .  .  .  .  Arrow: -
    25  .  .  .  .  .  .  .  .  .  Dir: 3
    26  .  .  .  .  .  .  .  .  .  Value: *ast.Ident {
    27  .  .  .  .  .  .  .  .  .  .  NamePos: -
    28  .  .  .  .  .  .  .  .  .  .  Name: "int"
    29  .  .  .  .  .  .  .  .  .  }
    30  .  .  .  .  .  .  .  .  }
    31  .  .  .  .  .  .  .  }
    32  .  .  .  .  .  .  }
    33  .  .  .  .  .  }
    34  .  .  .  .  .  Arrow: -
    35  .  .  .  .  .  Value: *ast.Ident {
    36  .  .  .  .  .  .  NamePos: -
    37  .  .  .  .  .  .  Name: "sum"
    38  .  .  .  .  .  .  Obj: *ast.Object {
    39  .  .  .  .  .  .  .  Kind: var
    40  .  .  .  .  .  .  .  Name: "sum"
    41  .  .  .  .  .  .  .  Decl: *ast.AssignStmt {
    42  .  .  .  .  .  .  .  .  Lhs: []ast.Expr (len = 1) {
    43  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
    44  .  .  .  .  .  .  .  .  .  .  NamePos: -
    45  .  .  .  .  .  .  .  .  .  .  Name: "sum"
    46  .  .  .  .  .  .  .  .  .  .  Obj: *(obj @ 38)
    47  .  .  .  .  .  .  .  .  .  }
    48  .  .  .  .  .  .  .  .  }
    49  .  .  .  .  .  .  .  .  TokPos: -
    50  .  .  .  .  .  .  .  .  Tok: :=
    51  .  .  .  .  .  .  .  .  Rhs: []ast.Expr (len = 1) {
    52  .  .  .  .  .  .  .  .  .  0: *ast.BasicLit {
    53  .  .  .  .  .  .  .  .  .  .  ValuePos: -
    54  .  .  .  .  .  .  .  .  .  .  Kind: INT
    55  .  .  .  .  .  .  .  .  .  .  Value: "0"
    56  .  .  .  .  .  .  .  .  .  }
    57  .  .  .  .  .  .  .  .  }
    58  .  .  .  .  .  .  .  }
    59  .  .  .  .  .  .  }
    60  .  .  .  .  .  }
    61  .  .  .  .  }
    62  .  .  .  .  Colon: -
    63  .  .  .  .  Body: []ast.Stmt (len = 1) {
    64  .  .  .  .  .  0: *ast.BlockStmt {
    65  .  .  .  .  .  .  Lbrace: -
    66  .  .  .  .  .  .  List: []ast.Stmt (len = 1) {
    67  .  .  .  .  .  .  .  0: *ast.AssignStmt {
    68  .  .  .  .  .  .  .  .  Lhs: []ast.Expr (len = 1) {
    69  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
    70  .  .  .  .  .  .  .  .  .  .  NamePos: -
    71  .  .  .  .  .  .  .  .  .  .  Name: "_"
    72  .  .  .  .  .  .  .  .  .  }
    73  .  .  .  .  .  .  .  .  }
    74  .  .  .  .  .  .  .  .  TokPos: -
    75  .  .  .  .  .  .  .  .  Tok: =
    76  .  .  .  .  .  .  .  .  Rhs: []ast.Expr (len = 1) {
    77  .  .  .  .  .  .  .  .  .  0: *ast.BinaryExpr {
    78  .  .  .  .  .  .  .  .  .  .  X: *ast.BasicLit {
    79  .  .  .  .  .  .  .  .  .  .  .  ValuePos: -
    80  .  .  .  .  .  .  .  .  .  .  .  Kind: INT
    81  .  .  .  .  .  .  .  .  .  .  .  Value: "5"
    82  .  .  .  .  .  .  .  .  .  .  }
    83  .  .  .  .  .  .  .  .  .  .  OpPos: -
    84  .  .  .  .  .  .  .  .  .  .  Op: *
    85  .  .  .  .  .  .  .  .  .  .  Y: *ast.BasicLit {
    86  .  .  .  .  .  .  .  .  .  .  .  ValuePos: -
    87  .  .  .  .  .  .  .  .  .  .  .  Kind: INT
    88  .  .  .  .  .  .  .  .  .  .  .  Value: "6"
    89  .  .  .  .  .  .  .  .  .  .  }
    90  .  .  .  .  .  .  .  .  .  }
    91  .  .  .  .  .  .  .  .  }
    92  .  .  .  .  .  .  .  }
    93  .  .  .  .  .  .  }
    94  .  .  .  .  .  .  Rbrace: -
    95  .  .  .  .  .  }
    96  .  .  .  .  }
    97  .  .  .  }
    98  .  .  .  1: *ast.CommClause {
    99  .  .  .  .  Case: -
   100  .  .  .  .  Colon: -
   101  .  .  .  .  Body: []ast.Stmt (len = 1) {
   102  .  .  .  .  .  0: *ast.BlockStmt {
   103  .  .  .  .  .  .  Lbrace: -
   104  .  .  .  .  .  .  List: []ast.Stmt (len = 1) {
   105  .  .  .  .  .  .  .  0: *ast.AssignStmt {
   106  .  .  .  .  .  .  .  .  Lhs: []ast.Expr (len = 1) {
   107  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
   108  .  .  .  .  .  .  .  .  .  .  NamePos: -
   109  .  .  .  .  .  .  .  .  .  .  Name: "_"
   110  .  .  .  .  .  .  .  .  .  }
   111  .  .  .  .  .  .  .  .  }
   112  .  .  .  .  .  .  .  .  TokPos: -
   113  .  .  .  .  .  .  .  .  Tok: =
   114  .  .  .  .  .  .  .  .  Rhs: []ast.Expr (len = 1) {
   115  .  .  .  .  .  .  .  .  .  0: *ast.BinaryExpr {
   116  .  .  .  .  .  .  .  .  .  .  X: *ast.BasicLit {
   117  .  .  .  .  .  .  .  .  .  .  .  ValuePos: -
   118  .  .  .  .  .  .  .  .  .  .  .  Kind: INT
   119  .  .  .  .  .  .  .  .  .  .  .  Value: "4"
   120  .  .  .  .  .  .  .  .  .  .  }
   121  .  .  .  .  .  .  .  .  .  .  OpPos: -
   122  .  .  .  .  .  .  .  .  .  .  Op: *
   123  .  .  .  .  .  .  .  .  .  .  Y: *ast.BasicLit {
   124  .  .  .  .  .  .  .  .  .  .  .  ValuePos: -
   125  .  .  .  .  .  .  .  .  .  .  .  Kind: INT
   126  .  .  .  .  .  .  .  .  .  .  .  Value: "7"
   127  .  .  .  .  .  .  .  .  .  .  }
   128  .  .  .  .  .  .  .  .  .  }
   129  .  .  .  .  .  .  .  .  }
   130  .  .  .  .  .  .  .  }
   131  .  .  .  .  .  .  }
   132  .  .  .  .  .  .  Rbrace: -
   133  .  .  .  .  .  }
   134  .  .  .  .  }
   135  .  .  .  }
   136  .  .  }
   137  .  .  Rbrace: -
   138  .  }
   139  }


Printing AST tree:
- fset := token.NewFileSet()
-  ast.Print(fset, node)
*/
import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"reflect"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "channelcheck",
	Doc:  "reports channel blocking issues",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {

	for _, file := range pass.Files {
		var seenPositions = make(map[token.Pos]bool)

		ast.Inspect(file, func(node ast.Node) bool {
			switch n := node.(type) {
			// Fails open by design. Will
			case *ast.SelectStmt: // Select statement for channel matching

				channelSendFound, defaultOrTimeout, seenPositionsLocal := processSelect(pass, *n)
				/*
					If we found a 'SendStmt' alongside a default or a timer, then it's safe.
					If NOT found, this case will be covered and added as a linting error.
				*/
				if defaultOrTimeout && channelSendFound {
					// Add local Pos to global structure for later
					for key, value := range seenPositionsLocal {
						seenPositions[key] = value
					}
				}

			// Most of the work is done in the previous case statement.
			case *ast.SendStmt:

				// If the SendStmt was NOT found within a Select clause, then add a linter error.
				tokenId := n.Pos()
				if _, ok := seenPositions[tokenId]; !ok {
					pass.Reportf(tokenId, "channel send without default or timer - consider adding default or timeout case %q", render(pass.Fset, n))
				}

				return true
			case *ast.CallExpr:
				// Channel creation that's unbuffered
				didCreateChannelWithoutBuffering := checkChannelCreation(n)
				if didCreateChannelWithoutBuffering {
					pass.Reportf(n.Pos(), "unbuffered channel creation detected - consider specifying buffer size %q", render(pass.Fset, n))
				}
				return true

			default:
				return true // Continue traversing for other node types
			}

			return true
		})
	}

	return nil, nil
}

/*
Check the Select statement so see if it has the following:
- A SendStat (c <- data)
- Fallback case:
  - Default (nil)
  - Timeout
  - Ticker

NOTE: Channel function definitions do NOT include whether a channel is buffered or not.
So, it's impossible to determine whether SELECT statement is okay or not based upon that
criteria. As a result, if there's a 'Send' to a channel without fallback cases,
we must report it.
*/
func processSelect(pass *analysis.Pass, n ast.SelectStmt) (bool, bool, map[token.Pos]bool) {
	var seenPositionsLocal = make(map[token.Pos]bool)

	channelSendFound := false
	defaultOrTimeout := false
	for _, commClause := range n.Body.List { // Iterate through each clause in a select statement

		commClause, ok := commClause.(*ast.CommClause)
		if !ok {
			continue // Skip if not a CommClause (e.g., a declaration inside the select)
		}

		if reflect.TypeOf(commClause.Comm) == nil { // default case
			defaultOrTimeout = true
			continue
		}

		// From test cases, this seems sufficient.
		if sendNode, ok := commClause.Comm.(*ast.SendStmt); ok {
			// Handle nested select statement if needed
			channelSendFound = true
			seenPositionsLocal[sendNode.Pos()] = true
			continue
		}

		// callNode := findNode(pass, commClause.Comm)
		// if callNode != nil {
		// 	defaultOrTimeout = true
		// 	continue
		// }

		foundTimeout := findNodeTimeout(pass, commClause.Comm)
		if foundTimeout {
			defaultOrTimeout = true
		}
		/*
			TODO search for 'time.After' and 'ticker' packages.
			Simple example:
				0  *ast.CommClause {
				1  .  Case: -
				2  .  Comm: *ast.ExprStmt {
				3  .  .  X: *ast.UnaryExpr {
				4  .  .  .  OpPos: -
				5  .  .  .  Op: <-
				6  .  .  .  X: *ast.CallExpr {
				7  .  .  .  .  Fun: *ast.SelectorExpr {
				8  .  .  .  .  .  X: *ast.Ident {
				9  .  .  .  .  .  .  NamePos: -
				10  .  .  .  .  .  .  Name: "time"
				11  .  .  .  .  .  }
				12  .  .  .  .  .  Sel: *ast.Ident {
				13  .  .  .  .  .  .  NamePos: -
				14  .  .  .  .  .  .  Name: "After"
				15  .  .  .  .  .  }
				16  .  .  .  .  }
			Something more complicated with recursive searching may be necessary on the '
		*/

		//fset := token.NewFileSet() //
		//ast.Print(fset, commClause)
	}

	return channelSendFound, defaultOrTimeout, seenPositionsLocal

}

// findNode recursively searches the AST node for a node of the specified type
// using a depth-first approach.
func findNodeTimeout(pass *analysis.Pass, node ast.Node) bool {

	foundTimeout := false
	ast.Inspect(node, func(n ast.Node) bool {
		if foundTimeout == true { // Once we've found what we're looking for, short circuit the call
			return false
		}
		if n == nil {
			return false // Stop traversal at nil nodes
		}

		call, ok := n.(*ast.CallExpr)
		if ok {
			timeoutCall := isTimeAfter(pass, call)
			if timeoutCall {
				foundTimeout = true
			}
		}

		return true // Continue traversal to child nodes
	})

	return foundTimeout
}

// Is this too strict? Could be?
func isTimeAfter(pass *analysis.Pass, expr ast.Expr) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}

	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	// 1. Check the selector name
	if sel.Sel.Name != "After" {
		return false
	}

	// 2. Check if the receiver ("X") is an identifier
	id, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}

	// 3. CRUCIALLY: Use type information to verify it's the "time" package.
	obj := pass.TypesInfo.Uses[id]
	if obj == nil {
		return false // Identifier not found in type info
	}

	pkgName, ok := obj.(*types.PkgName)
	if !ok {
		return false // Not a package name
	}

	return pkgName.Imported().Path() == "time"
}

func checkChannelCreation(node *ast.CallExpr) bool {
	fun, ok := node.Fun.(*ast.Ident)
	if !ok || fun == nil || fun.Name != "make" {
		return false
	}

	if len(node.Args) > 0 {
		if chanType, ok := node.Args[0].(*ast.ChanType); ok && chanType != nil {
			// Check if buffer size is specified
			if len(node.Args) == 1 {
				return true
			}
		}
	}

	return false
}

// render returns the pretty-print of the given node
func render(fset *token.FileSet, x interface{}) string {
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, x); err != nil {
		panic(err)
	}
	return buf.String()
}
