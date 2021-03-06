package html5_test

import (
	. "github.com/bytesparadise/libasciidoc/testsupport"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

var _ = Describe("file inclusions", func() {

	It("include adoc file with leveloffset attribute", func() {
		source := `= Master Document

preamble

include::../../../test/includes/chapter-a.adoc[leveloffset=+1]`
		expected := `<div id="preamble">
<div class="sectionbody">
<div class="paragraph">
<p>preamble</p>
</div>
</div>
</div>
<div class="sect1">
<h2 id="_chapter_a">Chapter A</h2>
<div class="sectionbody">
<div class="paragraph">
<p>content</p>
</div>
</div>
</div>`
		Expect(source).To(RenderHTML5Element(expected))
	})

	It("include non adoc file", func() {
		source := `= Master Document

preamble

include::../../../test/includes/hello_world.go.txt[]`
		expected := `<div class="paragraph">
<p>preamble</p>
</div>
<div class="paragraph">
<p>package includes</p>
</div>
<div class="paragraph">
<p>import &#34;fmt&#34;</p>
</div>
<div class="paragraph">
<p>func helloworld() {
	fmt.Println(&#34;hello, world!&#34;)
}</p>
</div>`
		Expect(source).To(RenderHTML5Element(expected))
	})

	It("include 2 files", func() {
		source := `= Master Document

preamble

include::../../../test/includes/grandchild-include.adoc[]

include::../../../test/includes/hello_world.go.txt[]`
		expected := `<div class="paragraph">
<p>preamble</p>
</div>
<div class="paragraph">
<p>first line of grandchild</p>
</div>
<div class="paragraph">
<p>last line of grandchild</p>
</div>
<div class="paragraph">
<p>package includes</p>
</div>
<div class="paragraph">
<p>import &#34;fmt&#34;</p>
</div>
<div class="paragraph">
<p>func helloworld() {
	fmt.Println(&#34;hello, world!&#34;)
}</p>
</div>`
		Expect(source).To(RenderHTML5Element(expected))
	})

	It("include file and append following elements in included section", func() {
		source := `a first paragraph

include::../../../test/includes/chapter-a.adoc[leveloffset=+1]

a second paragraph

a third paragraph`
		expected := `<div class="paragraph">
<p>a first paragraph</p>
</div>
<div class="sect1">
<h2 id="_chapter_a">Chapter A</h2>
<div class="sectionbody">
<div class="paragraph">
<p>content</p>
</div>
<div class="paragraph">
<p>a second paragraph</p>
</div>
<div class="paragraph">
<p>a third paragraph</p>
</div>
</div>
</div>`
		Expect(source).To(RenderHTML5Element(expected))
	})

	Context("file inclusion in delimited blocks", func() {

		Context("adoc file inclusion in delimited blocks", func() {

			It("should include adoc file within listing block", func() {
				source := `= Master Document

preamble

----
include::../../../test/includes/chapter-a.adoc[]
----`
				expected := `<div class="paragraph">
<p>preamble</p>
</div>
<div class="listingblock">
<div class="content">
<pre>= Chapter A

content</pre>
</div>
</div>`
				Expect(source).To(RenderHTML5Element(expected))
			})

			It("should include adoc file within fenced block", func() {
				source := "```\n" +
					"include::../../../test/includes/chapter-a.adoc[]\n" +
					"```"
				expected := `<div class="listingblock">
<div class="content">
<pre class="highlight"><code>= Chapter A

content</code></pre>
</div>
</div>`
				Expect(source).To(RenderHTML5Element(expected))
			})

			It("should include adoc file within example block", func() {
				source := `====
include::../../../test/includes/chapter-a.adoc[]
====`
				expected := `<div class="exampleblock">
<div class="content">
<div class="paragraph">
<p>= Chapter A</p>
</div>
<div class="paragraph">
<p>content</p>
</div>
</div>
</div>`
				Expect(source).To(RenderHTML5Element(expected))
			})

			It("should include adoc file within quote block", func() {
				source := `____
include::../../../test/includes/chapter-a.adoc[]
____`
				expected := `<div class="quoteblock">
<blockquote>
<div class="paragraph">
<p>= Chapter A</p>
</div>
<div class="paragraph">
<p>content</p>
</div>
</blockquote>
</div>`
				Expect(source).To(RenderHTML5Element(expected))
			})

			It("should include adoc file within verse block", func() {
				source := `[verse]
____
include::../../../test/includes/chapter-a.adoc[]
____`
				expected := `<div class="verseblock">
<pre class="content">= Chapter A

content</pre>
</div>`
				Expect(source).To(RenderHTML5Element(expected))
			})

			It("should include adoc file within sidebar block", func() {
				source := `****
include::../../../test/includes/chapter-a.adoc[]
****`
				expected := `<div class="sidebarblock">
<div class="content">
<div class="paragraph">
<p>= Chapter A</p>
</div>
<div class="paragraph">
<p>content</p>
</div>
</div>
</div>`
				Expect(source).To(RenderHTML5Element(expected))
			})

			It("should include adoc file within passthrough block", func() {
				Skip("missing support for passthrough blocks")
				source := `++++
include::../../../test/includes/chapter-a.adoc[]
++++`
				expected := ``
				Expect(source).To(RenderHTML5Element(expected))
			})
		})

		Context("other file inclusion in delimited blocks", func() {

			It("should include go file within listing block", func() {
				source := `= Master Document

preamble

----
include::../../../test/includes/hello_world.go.txt[]
----`
				expected := `<div class="paragraph">
<p>preamble</p>
</div>
<div class="listingblock">
<div class="content">
<pre>package includes

import &#34;fmt&#34;

func helloworld() {
	fmt.Println(&#34;hello, world!&#34;)
}</pre>
</div>
</div>`
				Expect(source).To(RenderHTML5Element(expected))
			})

			It("should include go file within fenced block", func() {
				source := "```\n" +
					"include::../../../test/includes/hello_world.go.txt[]\n" +
					"```"
				expected := `<div class="listingblock">
<div class="content">
<pre class="highlight"><code>package includes

import "fmt"

func helloworld() {
	fmt.Println("hello, world!")
}</code></pre>
</div>
</div>`
				Expect(source).To(RenderHTML5Element(expected))
			})

			It("should include go file within example block", func() {
				source := `====
include::../../../test/includes/hello_world.go.txt[]
====`
				expected := `<div class="exampleblock">
<div class="content">
<div class="paragraph">
<p>package includes</p>
</div>
<div class="paragraph">
<p>import &#34;fmt&#34;</p>
</div>
<div class="paragraph">
<p>func helloworld() {
	fmt.Println(&#34;hello, world!&#34;)
}</p>
</div>
</div>
</div>`
				Expect(source).To(RenderHTML5Element(expected))
			})

			It("should include go file within quote block", func() {
				source := `____
include::../../../test/includes/hello_world.go.txt[]
____`
				expected := `<div class="quoteblock">
<blockquote>
<div class="paragraph">
<p>package includes</p>
</div>
<div class="paragraph">
<p>import &#34;fmt&#34;</p>
</div>
<div class="paragraph">
<p>func helloworld() {
	fmt.Println(&#34;hello, world!&#34;)
}</p>
</div>
</blockquote>
</div>`
				Expect(source).To(RenderHTML5Element(expected))
			})

			It("should include go file within verse block", func() {
				source := `[verse]
____
include::../../../test/includes/hello_world.go.txt[]
____`
				expected := `<div class="verseblock">
<pre class="content">package includes

import &#34;fmt&#34;

func helloworld() {
	fmt.Println(&#34;hello, world!&#34;)
}</pre>
</div>`
				Expect(source).To(RenderHTML5Element(expected))
			})

			It("should include go file within sidebar block", func() {
				source := `****
include::../../../test/includes/hello_world.go.txt[]
****`
				expected := `<div class="sidebarblock">
<div class="content">
<div class="paragraph">
<p>package includes</p>
</div>
<div class="paragraph">
<p>import &#34;fmt&#34;</p>
</div>
<div class="paragraph">
<p>func helloworld() {
	fmt.Println(&#34;hello, world!&#34;)
}</p>
</div>
</div>
</div>`
				Expect(source).To(RenderHTML5Element(expected))
			})
		})
	})

	Context("file inclusions with line range", func() {

		Context("file inclusions as paragraph with line range", func() {

			It("should include single line as paragraph", func() {
				source := `include::../../../test/includes/hello_world.go.txt[lines=1]`
				expected := `<div class="paragraph">
<p>package includes</p>
</div>`
				Expect(source).To(RenderHTML5Element(expected))
			})

			It("should include multiple lines as paragraph", func() {
				source := `include::../../../test/includes/hello_world.go.txt[lines=5..7]`
				expected := `<div class="paragraph">
<p>func helloworld() {
	fmt.Println(&#34;hello, world!&#34;)
}</p>
</div>`
				Expect(source).To(RenderHTML5Element(expected))
			})

			It("should include multiple ranges as paragraph", func() {
				source := `include::../../../test/includes/hello_world.go.txt[lines=1..2;5..7]`
				expected := `<div class="paragraph">
<p>package includes</p>
</div>
<div class="paragraph">
<p>func helloworld() {
	fmt.Println(&#34;hello, world!&#34;)
}</p>
</div>`
				Expect(source).To(RenderHTML5Element(expected))
			})
		})

		Context("file inclusions in listing blocks with line range", func() {

			It("should include single line in listing block", func() {
				source := `----
include::../../../test/includes/hello_world.go.txt[lines=1]
----`
				expected := `<div class="listingblock">
<div class="content">
<pre>package includes</pre>
</div>
</div>`
				Expect(source).To(RenderHTML5Element(expected))
			})

			It("should include multiple lines in listing block", func() {
				source := `----
include::../../../test/includes/hello_world.go.txt[lines=5..7]
----`
				expected := `<div class="listingblock">
<div class="content">
<pre>func helloworld() {
	fmt.Println(&#34;hello, world!&#34;)
}</pre>
</div>
</div>`
				Expect(source).To(RenderHTML5Element(expected))
			})

			It("should include multiple ranges in listing block", func() {
				source := `----
include::../../../test/includes/hello_world.go.txt[lines=1..2;5..7]
----`
				expected := `<div class="listingblock">
<div class="content">
<pre>package includes

func helloworld() {
	fmt.Println(&#34;hello, world!&#34;)
}</pre>
</div>
</div>`
				Expect(source).To(RenderHTML5Element(expected))
			})
		})
	})

	Context("file inclusions with tag ranges", func() {

		It("file inclusion with single tag", func() {
			source := `include::../../../test/includes/tag-include.adoc[tag=section]`
			expected := `<div class="sect1">
<h2 id="_section_1">Section 1</h2>
<div class="sectionbody">
</div>
</div>`
			Expect(source).To(RenderHTML5Element(expected))
		})

		It("file inclusion with surrounding tag", func() {
			source := `include::../../../test/includes/tag-include.adoc[tag=doc]`
			expected := `<div class="sect1">
<h2 id="_section_1">Section 1</h2>
<div class="sectionbody">
<div class="paragraph">
<p>content</p>
</div>
</div>
</div>`
			Expect(source).To(RenderHTML5Element(expected))
		})

		It("file inclusion with unclosed tag", func() {
			console, reset := ConfigureLogger()
			defer reset()
			source := `include::../../../test/includes/tag-include.adoc[tag=unclosed]`
			expected := `<div class="paragraph">
<p>content</p>
</div>
<div class="paragraph">
<p>end</p>
</div>`
			Expect(source).To(RenderHTML5Element(expected))
			// verify error in logs
			Expect(console).To(
				ContainMessageWithLevel(
					log.ErrorLevel,
					"detected unclosed tag 'unclosed' starting at line 6 of include file: ../../../test/includes/tag-include.adoc",
				))
		})

		It("file inclusion with unknown tag", func() {
			console, reset := ConfigureLogger()
			defer reset()
			source := `include::../../../test/includes/tag-include.adoc[tag=unknown]`
			expected := ``
			Expect(source).To(RenderHTML5Element(expected))
			// verify error in logs
			Expect(console).To(
				ContainMessageWithLevel(
					log.ErrorLevel,
					"tag 'unknown' not found in include file: ../../../test/includes/tag-include.adoc",
				))
		})

		It("file inclusion with no tag", func() {
			source := `include::../../../test/includes/tag-include.adoc[]`
			expected := `<div class="sect1">
<h2 id="_section_1">Section 1</h2>
<div class="sectionbody">
<div class="paragraph">
<p>content</p>
</div>
<div class="paragraph">
<p>end</p>
</div>
</div>
</div>`
			Expect(source).To(RenderHTML5Element(expected))
		})
	})

	Context("recursive file inclusions", func() {

		It("should include child and grandchild content in paragraphs", func() {
			source := `include::../../../test/includes/parent-include.adoc[]`
			expected := `<div class="paragraph">
<p>first line of parent</p>
</div>
<div class="paragraph">
<p>first line of child</p>
</div>
<div class="paragraph">
<p>first line of grandchild</p>
</div>
<div class="paragraph">
<p>last line of grandchild</p>
</div>
<div class="paragraph">
<p>last line of child</p>
</div>
<div class="paragraph">
<p>last line of parent</p>
</div>`
			Expect(source).To(RenderHTML5Element(expected))
		})

		It("should include child and grandchild content in listing block", func() {
			source := `----
include::../../../test/includes/parent-include.adoc[]
----`
			expected := `<div class="listingblock">
<div class="content">
<pre>first line of parent

first line of child

first line of grandchild

last line of grandchild

last line of child

last line of parent</pre>
</div>
</div>`
			Expect(source).To(RenderHTML5Element(expected))
		})
	})

	Context("inclusion with attribute in path", func() {

		It("should resolve path with attribute in standalone block", func() {
			source := `:includedir: ../../../test/includes
			
include::{includedir}/grandchild-include.adoc[]`
			expected := `<div class="paragraph">
<p>first line of grandchild</p>
</div>
<div class="paragraph">
<p>last line of grandchild</p>
</div>`
			Expect(source).To(RenderHTML5Element(expected))
		})

		It("should resolve path with attribute in delimited block", func() {
			source := `:includedir: ../../../test/includes

----
include::{includedir}/grandchild-include.adoc[]
----`
			expected := `<div class="listingblock">
<div class="content">
<pre>first line of grandchild

last line of grandchild</pre>
</div>
</div>`
			Expect(source).To(RenderHTML5Element(expected))
		})
	})

	Context("missing file to include", func() {

		Context("in standalone block", func() {

			It("should replace with string element if file is missing", func() {
				// setup logger to write in a buffer so we can check the output
				console, reset := ConfigureLogger()
				defer reset()

				source := `include::../../../test/includes/unknown.adoc[leveloffset=+1]`
				expected := `<div class="paragraph">
<p>Unresolved directive in test.adoc - include::../../../test/includes/unknown.adoc[leveloffset=&#43;1]</p>
</div>`
				Expect(source).To(RenderHTML5Element(expected))
				// verify error in logs
				Expect(console).To(
					ContainMessageWithLevel(
						log.ErrorLevel,
						"failed to include '../../../test/includes/unknown.adoc'",
					))
			})

			It("should replace with string element if file with attribute in path is not resolved", func() {
				// setup logger to write in a buffer so we can check the output
				console, reset := ConfigureLogger()
				defer reset()

				source := `include::{includedir}/unknown.adoc[leveloffset=+1]`
				expected := `<div class="paragraph">
<p>Unresolved directive in test.adoc - include::{includedir}/unknown.adoc[leveloffset=&#43;1]</p>
</div>`
				Expect(source).To(RenderHTML5Element(expected))
				// verify error in logs
				Expect(console).To(
					ContainMessageWithLevel(
						log.ErrorLevel,
						"failed to include '{includedir}/unknown.adoc'",
					))
			})
		})

		Context("in listing block", func() {

			It("should replace with string element if file is missing", func() {
				// setup logger to write in a buffer so we can check the output
				console, reset := ConfigureLogger()
				defer reset()

				source := `----
include::../../../test/includes/unknown.adoc[leveloffset=+1]
----`
				expected := `<div class="listingblock">
<div class="content">
<pre>Unresolved directive in test.adoc - include::../../../test/includes/unknown.adoc[leveloffset=+1]</pre>
</div>
</div>`
				Expect(source).To(RenderHTML5Element(expected))
				// verify error in logs
				Expect(console).To(
					ContainMessageWithLevel(
						log.ErrorLevel,
						"failed to include '../../../test/includes/unknown.adoc'",
					))
			})

			It("should replace with string element if file with attribute in path is not resolved", func() {
				// setup logger to write in a buffer so we can check the output
				console, reset := ConfigureLogger()
				defer reset()

				source := `----
include::{includedir}/unknown.adoc[leveloffset=+1]
----`
				expected := `<div class="listingblock">
<div class="content">
<pre>Unresolved directive in test.adoc - include::{includedir}/unknown.adoc[leveloffset=+1]</pre>
</div>
</div>`
				Expect(source).To(RenderHTML5Element(expected))
				// verify error in logs
				Expect(console).To(
					ContainMessageWithLevel(
						log.ErrorLevel,
						"failed to include '{includedir}/unknown.adoc'",
					))
			})
		})
	})
})
