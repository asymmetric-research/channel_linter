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

	Integrate with the golangci-lint system.
	- https://github.com/uber-go/nilaway/blob/ba14292918d814eeaea4de62da2ad0daae92f8b0/README.md
	- https://golangci-lint.run/plugins/module-plugins/

	Inspired by John's PoC implementation: https://github.com/johnsaigle/channelcheck

	Resources:
	- https://developer20.com/custom-go-linter/
	- https://arslan.io/2019/06/13/using-go-analysis-to-write-a-custom-linter/

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
fset := token.NewFileSet()
ast.Print(fset, node)
*/
import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/constant"
	"go/printer"
	"go/token"
	"go/types"
	"reflect"

	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"
)

type ChannelCheckPlugin struct {
	settings Settings
}

// Settings holds the configuration for the channelcheck linter.
type Settings struct {
	CheckUnbufferedChannels bool   // Enable/disable checking for unbuffered channel creation.
	CheckBufferAmount       uint64 // The amount that can be in a buffer. 0 means don't do this check.
	CheckBlockingSends      bool   // Enable/disable checking for blocking sends without default/timeout.
}

var Analyzer = &analysis.Analyzer{
	Name:  "channelcheck",
	Doc:   "reports channel blocking issues",
	Run:   run,
	Flags: flagSet,
}

// Flags for the analyzer
var flagSet flag.FlagSet

// Global structure to store the variables in
var settings Settings

func New(settings_new any) (register.LinterPlugin, error) {
	s, err := register.DecodeSettings[Settings](settings_new)
	if err != nil {
		return nil, err
	}
	settings.CheckBlockingSends = s.CheckBlockingSends
	settings.CheckBufferAmount = s.CheckBufferAmount
	settings.CheckUnbufferedChannels = s.CheckUnbufferedChannels

	return &ChannelCheckPlugin{settings: s}, nil
}

func (f *ChannelCheckPlugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{
		{
			Name:  "channelcheck",
			Doc:   "reports channel blocking issues",
			Run:   run,
			Flags: flagSet,
		},
	}, nil
}

// Initialize the flags from the golangci-lint
func init() {

	flagSet.BoolVar(&settings.CheckUnbufferedChannels, "unbuffered", false, "Check for unbuffered channel creation")
	flagSet.BoolVar(&settings.CheckBlockingSends, "blocking", true, "Check for blocking sends without default/timeout")
	flagSet.Uint64Var(&settings.CheckBufferAmount, "bufferMax", 0, "Check for maximum length of channel buffer being exceeded")
	register.Plugin("channelcheck", New)
	return
}

func (f *ChannelCheckPlugin) GetLoadMode() string {
	return register.LoadModeSyntax
}

func run(pass *analysis.Pass) (interface{}, error) {

	for _, file := range pass.Files {
		var seenPositions = make(map[token.Pos]bool)

		ast.Inspect(file, func(node ast.Node) bool {
			switch n := node.(type) {
			// Fails open by design. Will
			case *ast.SelectStmt: // Select statement for channel matching

				if settings.CheckBlockingSends == false {
					break
				}
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

				if settings.CheckBlockingSends == false {
					break
				}
				// If the SendStmt was NOT found within a Select clause, then add a linter error.
				tokenId := n.Pos()
				if _, ok := seenPositions[tokenId]; !ok {
					pass.Reportf(tokenId, "channel send without default or timer - consider adding default or timeout case %q", render(pass.Fset, n))
				}

				return true
			case *ast.CallExpr:
				// Channel creation that's unbuffered
				didCreateChannelWithoutBuffering, bufferAmount := checkChannelCreation(pass, n)
				if didCreateChannelWithoutBuffering && settings.CheckUnbufferedChannels {
					pass.Reportf(n.Pos(), "unbuffered channel creation detected - consider specifying buffer size %q", render(pass.Fset, n))
				}

				if settings.CheckBufferAmount > 0 && bufferAmount == -1 {
					pass.Reportf(n.Pos(), "channel buffer size set to 0 %q", render(pass.Fset, n))
				} else if settings.CheckBufferAmount > 0 && uint64(bufferAmount) > settings.CheckBufferAmount {
					pass.Reportf(n.Pos(), "channel buffer size exceeds the specified limit %q", render(pass.Fset, n))
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

		// From test cases, this seems sufficient.
		if sendNode, ok := commClause.Comm.(*ast.SendStmt); ok {
			// Handle nested select statement if needed
			channelSendFound = true
			seenPositionsLocal[sendNode.Pos()] = true
			continue
		}

		if reflect.TypeOf(commClause.Comm) == nil { // 'default' case
			defaultOrTimeout = true
			continue
		}

		// Timeout receive call. If the type being checked is 'time.Time', this is assumed to be a timeout but isn't 100% accurate.
		foundTimeout := findNodeTimeout(pass, commClause.Comm)
		if foundTimeout {
			defaultOrTimeout = true
		}
	}

	return channelSendFound, defaultOrTimeout, seenPositionsLocal
}

// findNode recursively searches the AST node for a node of the specified type
// using a depth-first approach.
func findNodeTimeout(pass *analysis.Pass, node ast.Node) bool {

	foundTimeout := false

	// All channel receives should be 'UnaryExpr' types with an Op of '<-'. Checking for this to reduce computations.
	node, ok := node.(*ast.ExprStmt)
	if !ok {
		return false
	}

	// Found all receive expressions only
	nodeExpr, ok := node.(*ast.ExprStmt).X.(*ast.UnaryExpr)
	if !ok || nodeExpr.Op != token.ARROW {
		return false
	}

	// If it's a TICKER type
	if isTimeReturnType(pass, nodeExpr) {
		return true
	}

	// Need for 'timeAfter' call.
	// ast.Inspect(nodeExpr, func(n ast.Node) bool {
	// 	if foundTimeout == true { // Once we've found what we're looking for, short circuit the call
	// 		return false
	// 	}
	// 	if n == nil {
	// 		return false // Stop traversal at nil nodes
	// 	}

	// 	call, ok := n.(*ast.CallExpr)
	// 	if ok {
	// 		timeoutCall := isTimeAfter(pass, call)
	// 		if timeoutCall {
	// 			foundTimeout = true
	// 			return false
	// 		}
	// 	}

	// 	return true // Continue traversal to child nodes
	// })

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

	// 3. Use type information to verify it's the "time" package.
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

/*
If the channel receive is for 'time.Time' types (which many of the tickers and timeouts do),
then we assume it's safe.

Could be done easier if it was possible to parse ALL variants of time.After and tickers.
In reality, this is super hard to do because you need to deal with all possible situations of variable assignment and such. This is quick and simple, which I really like.

NOTE: This IS prone to false negatives if there is another channel sending time.Time for another reason.
This is such a great way to do this I'm okay with this false negative though.
*/
func isTimeReturnType(pass *analysis.Pass, expr ast.Expr) bool {
	typeOfExpr := pass.TypesInfo.TypeOf(expr) // time.Time is interesting here
	if typeOfExpr == nil {
		return false // Or report an error
	}

	if typeOfExpr.String() == "time.Time" {
		return true
	}

	return false
}

func checkChannelCreation(pass *analysis.Pass, node *ast.CallExpr) (bool, int64) {
	fun, ok := node.Fun.(*ast.Ident)
	if !ok || fun == nil || fun.Name != "make" {
		return false, 0
	}

	if len(node.Args) > 0 {
		if _, ok := node.Args[0].(*ast.ChanType); ok { // It's a channel
			if len(node.Args) == 1 {
				return true, 0 // Unbuffered channel
			}

			if len(node.Args) == 2 {
				// Evaluate the buffer size expression
				bufferSize, err := evalBufferSize(pass, node.Args[1])
				if err != nil {
					return true, 0 // Or another error indicator
				}

				if bufferSize == 0 {
					return false, -1
				}
				return false, int64(bufferSize)
			}
		}
	}

	return false, 0
}

/*
Only supports a literal in the buffer size slot. Could expand to more complicated cases.
*/
func evalBufferSize(pass *analysis.Pass, expr ast.Expr) (uint64, error) {
	// 1. Check for integer literal (BasicLit)
	if lit, ok := expr.(*ast.BasicLit); ok && lit.Kind == token.INT {
		val := constant.ToInt(constant.MakeFromLiteral(lit.Value, lit.Kind, 0))
		if val.Kind() == constant.Int {
			bufferSize, exact := constant.Uint64Val(val)
			if !exact {
				return 0, fmt.Errorf("buffer size is too large")
			}
			return bufferSize, nil
		} else {
			return 0, fmt.Errorf("invalid buffer size type: %v", val.Kind()) // More specific error
		}
	}
	return 0, fmt.Errorf("invalid buffer size expression")
}

// render returns the pretty-print of the given node
func render(fset *token.FileSet, x interface{}) string {
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, x); err != nil {
		panic(err)
	}
	return buf.String()
}
