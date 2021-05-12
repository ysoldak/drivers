
clean:
	@rm -rf build
	@rm *.mk

FMT_PATHS = ./*.go ./examples/**/*.go

fmt-check:
	@unformatted=$$(gofmt -l $(FMT_PATHS)); [ -z "$$unformatted" ] && exit 0; echo "Unformatted:"; for fn in $$unformatted; do echo "  $$fn"; done; exit 1

./build/%.hex:
	@mkdir -p "$(@D)"
	tinygo build -size short -o $@ -target=$(notdir $(basename $@)) $(subst build,./examples,$(dir $@))
	@md5sum $@

TARGETS = arduino arduino-nano33 bluepill circuitplay-bluefruit circuitplay-express \
		digispark feather-m0 hifive1b itsybitsy-m0 microbit nucleo-f103rb nucleo-l432kc \
		p1am-100 pybadge pyportal trinket-m0 xiao
TARGET_TAGS = $(shell tinygo info $(target) | grep "build tags" | sed "s/build tags:\s*//" | tr -s ' ')
EXAMPLES = $(dir $(shell find ./examples -type f -name 'main.go'))
EXAMPLE_TAGS = $(subst // +build ,,$(shell head -1 $(example)main.go))
EXAMPLE_TARGET_FILE = $(subst examples,build,$(example)$(target).hex)
EXAMPLE_TARGET_FILES = $(if $(filter $(TARGET_TAGS),$(EXAMPLE_TAGS)),$(EXAMPLE_TARGET_FILE))
EXAMPLE_HEX_FILES = $(foreach example,$(EXAMPLES),$(foreach target,$(TARGETS), $(EXAMPLE_TARGET_FILES)))

smoke-test.mk: Makefile
	for example_hex_file in $(EXAMPLE_HEX_FILES); do \
		echo "smoke-test: $${example_hex_file}" >> smoke-test.mk; \
	done

-include smoke-test.mk

DRIVERS = $(wildcard */)
NOTESTS = build examples flash semihosting pcd8544 shiftregister st7789 microphone mcp3008 gps microbitmatrix \
		hcsr04 ssd1331 ws2812 thermistor apa102 easystepper ssd1351 ili9341 wifinina shifter hub75 \
		hd44780 buzzer ssd1306 espat l9110x st7735 bmi160 l293x dht keypad4x4 max72xx p1am tone tm1637 \
		pcf8563 mcp2515
TESTS = $(filter-out $(addsuffix /%,$(NOTESTS)),$(DRIVERS))

unit-test:
	@go test -v $(addprefix ./,$(TESTS)) 

test: clean fmt-check unit-test smoke-test
