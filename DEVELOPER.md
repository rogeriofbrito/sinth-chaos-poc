# developer

## Install mockery

```bash
brew install mockery
```

## Install ginkgo

```bash
go install github.com/onsi/ginkgo/v2/ginkgo
```

## Create new test suit

```bash
cd <test folder>
ginkgo generate <test suite name>
```

## Generate mocks

```bash
cd <root folder>
mockery
```
