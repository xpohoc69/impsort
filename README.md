# impsort

`go install github.com/xpohoc69/impsort@latest`

Sorts the imports in the project .go files.

Run from the root directory of the project or pass the path to the project as the first argument.

`impsort /home/john/go/src/gitlab.com/services/example`

Before:

```
import (
    "gitlab.mycompany.com/libs/golang/logger"
    "gitlab.mycompany.com/services/tinkoff"
    "go.temporal.io/sdk/worker"
    "log"
)
```

After:

```
import (
    "log"

    "gitlab.mycompany.com/services/tinkoff"

    "gitlab.mycompany.com/libs/golang/logger"

    "go.temporal.io/sdk/worker"
)
```
