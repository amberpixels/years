// export_test.go declares exported variables/functions from internal ones
// to be used in `_test.go` files.
package years

import "time"

var CoreAliases = coreAliasesFor(time.Sunday, defaultWorkdays())
