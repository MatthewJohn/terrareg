
import re

from markdown import Markdown
from markdown.treeprocessors import Treeprocessor
from markdown.preprocessors import Preprocessor
from markdown.htmlparser import HTMLExtractor
# Avoid overlap with markdown method
import markdown.extensions as markdown_extensions


class CustomMarkdown(Markdown):
    """Override Markdown class to accept file_name argument"""
    
    def __init__(self, file_name, *args, **kwargs):
        """Run super init method and store file name"""
        super(CustomMarkdown, self).__init__(*args, **kwargs)
        self.terrareg_file_name = file_name


class LinkAnchorReplacement(Treeprocessor):
    """Add IDs to headings and convert links to current filename"""

    @staticmethod
    def _convert_id(id):
        return f"terrareg-markdown-anchor-{id}"

    def run(self, root):
        """Replace IDs of anchorable elements and replace anchor links with correct ID"""
        escaped_file_name = re.escape(self.md.terrareg_file_name)
        link_re = re.compile(rf"^(?:(?:\.\/)?{escaped_file_name})?#(.*)$")

        # Iterate over links, replacing href with Terrareg links
        for link in root.findall('.//a'):
            link_match = link_re.match(link.attrib.get('href', ''))
            if link_match:
                link.attrib['href'] = f"#{self._convert_id(link_match.group(1))}"

        # Add ID to head h1, h2, etc.
        for tag_name in ['h1', 'h2', 'h3', 'h4', 'h5', 'h6']:
            for tag in root.findall(f'.//{tag_name}'):
                # Generate anchor from text and convert to ID
                if tag.text:
                    tag.attrib['id'] = self._convert_id(tag.text.lower().replace(' ', '-'))
        
        # Convert pre-existing name tags
        for tag in root.findall('.//a'):
            if name_attribute := tag.attrib.get('name'):
                tag.attrib['name'] = self._convert_id(name_attribute)


class HTMLExtractorWithAttribs(HTMLExtractor):

    def handle_starttag(self, tag, attrs):
        return super(HTMLExtractorWithAttribs, self).handle_starttag(tag, attrs)

    def handle_startendtag(self, tag, attrs):
        self.handle_empty_tag(self.get_starttag_text(), is_block=self.md.is_block_level(tag))


class HtmlBlockPreprocessorWithAttribs(Preprocessor):
    """Preprocessor to override HtmlBlockPreprocessor, using HTMLExtractorWithAttribs"""

    def run(self, lines):
        source = '\n'.join(lines)
        parser = HTMLExtractorWithAttribs(self.md)
        parser.feed(source)
        parser.close()
        return ''.join(parser.cleandoc).split('\n')



class TerraregMarkdownExtension(markdown_extensions.Extension):
    """Markdown extension for Terrareg."""

    def extendMarkdown(self, md):
        """Register processors"""
        # Replace raw HTML preprocessor
        md.preprocessors.register(HtmlBlockPreprocessorWithAttribs(md), 'html_block', 20)
        md.treeprocessors.register(LinkAnchorReplacement(md), 'linkanchorreplacement', 1)


def makeExtension(**kwargs):
    """Create instance of markdown extension"""
    return TerraregMarkdownExtension(**kwargs)


def markdown(text, **kwargs):
    """Replicate upstream markdown method to create instance of custom Markdown class"""
    md = CustomMarkdown(**kwargs)
    return md.convert(text)
