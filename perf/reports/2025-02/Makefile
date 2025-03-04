BASE_NAME = veraison-aws-perf
MARP_OPTS = --allow-local-files --theme theme.css --html

.DEFAULT: all

.PHONY: all pdf html pptx clean plots

PLOT_ROOT = ../../analysis/2025-02

all: pdf html pptx

pdf: $(BASE_NAME).pdf

html: $(BASE_NAME).html

plots:
	mkdir -p images/3p3v
	mkdir -p images/1p1v
	mkdir -p images/comp-3v1
	mkdir -p images/auto
	mkdir -p images/weights
	cp $(PLOT_ROOT)/aws-3prov-3verif/summary/*{png,html}  images/3p3v/
	cp $(PLOT_ROOT)/aws-3prov-3verif/timelines/*{png,html}  images/3p3v/
	cp $(PLOT_ROOT)/aws-1prov-1verif/summary/*{png,html}  images/1p1v/
	cp $(PLOT_ROOT)/aws-1prov-1verif/timelines/*{png,html}  images/1p1v/
	cp $(PLOT_ROOT)/comp-3v1/*{png,html}  images/comp-3v1/
	cp $(PLOT_ROOT)/aws-autoscaled/summary/*{png,html}  images/auto/
	cp $(PLOT_ROOT)/aws-autoscaled/timelines/*{png,html}  images/auto/
	cp $(PLOT_ROOT)/aws-store-weights/*{png,html}  images/weights/

# NOTE: in order to build
$(BASE_NAME).pdf: $(BASE_NAME).md plots
	marp $(BASE_NAME).md $(MARP_OPTS) --pdf --pdf-notes --pdf-outline

$(BASE_NAME).html: $(BASE_NAME).md plots
	marp $(BASE_NAME).md $(MARP_OPTS)

watch:
	marp . -p -w --server $(MARP_OPTS)

clean:
	$(RM) -f *.pdf *.html *.pptx images
