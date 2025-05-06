#Variables
BINARY_NAME=telecomtask
CMD_DIR=cmd
MAIN_FILE=$(CMD_DIR)/main.go
BUILD_DIR=bin
TEST_DIR=.
GO=go
GOFMT=gofmt
GOLINT=golint

GREEN=\033[0;32m
NC=\033[0m

.PHONY: build
build:
	@echo "${GREEN}Building the project...${NC}"
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)

.PHONY: run
run: build
	@echo "${GREEN}Running the application...${NC}"
	./$(BUILD_DIR)/$(BINARY_NAME)

.PHONY: test
test:
	@echo "${GREEN}Running tests...${NC}"
	$(GO) test -v $(TEST_DIR)/...

.PHONY: clean
clean:
	@echo "${GREEN}Cleaning up...${NC}"
	rm -rf $(BUILD_DIR)
	$(GO) clean

.PHONY: deps
deps:
	@echo "${GREEN}Installing dependencies...${NC}"
	$(GO) mod tidy
	$(GO) mod download