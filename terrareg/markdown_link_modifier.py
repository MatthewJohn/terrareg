from typing import List, Tuple, Sequence
import re
import xml.etree.ElementTree as etree

from markdown import Markdown
from markdown.treeprocessors import Treeprocessor
from markdown.preprocessors import Preprocessor
from markdown.htmlparser import HTMLExtractor
from markdown.postprocessors import Postprocessor
# Avoid overlap with markdown method
import markdown.extensions as markdown_extensions

NON_ALPHANUMERIC_REPLACEMENT_RE = re.compile(r"[^a-zA-Z0-9\-_]")
HYPHEN_REPLACEMENT_RE = re.compile(r"[\s\-]+")

def _convert_id(file_name, id):
    file_name = NON_ALPHANUMERIC_REPLACEMENT_RE.sub("", HYPHEN_REPLACEMENT_RE.sub("-", file_name))
    # If 'id' provided is not a string, return the original value
    if not isinstance(id, str):
        return None

    id = NON_ALPHANUMERIC_REPLACEMENT_RE.sub("", HYPHEN_REPLACEMENT_RE.sub("-", id))
    return f"terrareg-anchor-{file_name}-{id}"

def _get_anchor_from_href(file_name, href):
    """Get anchor tag from href"""
    # Escape filename for use in regex
    escaped_file_name = re.escape(file_name)
    # Extract anchor from link
    link_re = re.compile(rf"^(?:(?:\.\/)?{escaped_file_name})?#(.*)$")
    match = link_re.match(href)
    # If an anchor matched, convert the ID to terrareg anchor and return with anchor #
    if match:
        return '#' + _convert_id(file_name=file_name, id=match.group(1))
    # Otherwise, if no match was found,
    # return original ref
    return href

class CustomMarkdown(Markdown):
    """Override Markdown class to accept file_name argument"""
    
    def __init__(self, file_name, *args, **kwargs):
        """Run super init method and store file name"""
        super(CustomMarkdown, self).__init__(*args, **kwargs)
        self.terrareg_file_name = file_name


class LinkAnchorReplacement(Treeprocessor):
    """Add IDs to headings and convert links to current filename"""

    def run(self, root):
        """Replace IDs of anchor-able elements and replace anchor links with correct ID"""

        # Iterate over links, replacing href with Terrareg links
        for link in root.findall('.//a'):
            if href := link.attrib.get('href', None):
                link.attrib['href'] = _get_anchor_from_href(file_name=self.md.terrareg_file_name, href=href)

        # Add ID to head h1, h2, etc.
        for tag_name in ['h1', 'h2', 'h3', 'h4', 'h5', 'h6']:
            for tag in root.findall(f'.//{tag_name}'):
                # Generate anchor from text and convert to ID
                if tag.text:
                    tag.attrib['id'] = _convert_id(self.md.terrareg_file_name, tag.text.lower().replace(' ', '-'))


class ImageSourceCheck(Treeprocessor):
    """Check image source URLs to ensure that they are from external resources"""

    def run(self, root):
        """Remove img src tags that do use relative paths for images"""
        for link in root.findall('.//img'):
            if src := link.attrib.get('src', ''):
                # Delete src attribute, if it does not
                # start with http:// or https://.
                # Relative URLs within a repository will not work
                # and will be displayed as broken images, which
                # does not look nice.
                # Removing the 'src' attribute will show white space,
                # rather than a broken image icon
                if (not src.startswith('http://')) and (not src.startswith('https://')):
                    del link.attrib['src']


class HTMLExtractorWithAttribs(HTMLExtractor):
    """Custom HTMLExtractor override, replacing name attributes of embedded HTML"""

    def reset(self):
        """Reset/setup member variables."""
        self.__starttag_text = None
        return super().reset()

    def handle_starttag(self, tag, attrs):
        """Convert name attributes to use custom ID"""
        # Replace name attribute values with converted ID
        converted_attribute = False
        for itx, attr in enumerate(attrs):
            if attr[0] == 'name' or attr[0] == 'id':
                if result_id := _convert_id(self.md.terrareg_file_name, attr[1]):
                    attrs[itx] = (attr[0], result_id)
                    converted_attribute = True
            if attr[0] == 'href':
                attrs[itx] = (attr[0], _get_anchor_from_href(file_name=self.md.terrareg_file_name, href=attr[1]))
                converted_attribute = True

        # If any of the attributes have been modified,
        # update the starttag_text attribute, so that
        # the modified version of the start tag is returned
        # by get_starttag_text.
        # Otherwise, set to None, to call super method
        # to retain the original functionality (which is
        # likely the be the same code as below).
        if converted_attribute:
            attribute_string = " ".join([f'{attr[0]}="{attr[1]}"' for attr in attrs])
            self.__starttag_text = f"<{tag} {attribute_string}>"
        else:
            # The instance of this class is shared between elements,
            # if it is not reset, we will continue returning the custom
            # __starttag_text, as set above, for all subsequent elements.
            self.__starttag_text = None

        return super(HTMLExtractorWithAttribs, self).handle_starttag(tag, attrs)

    def get_starttag_text(self):
        """Return modified start tag text, if attribute modifications were made"""
        if self.__starttag_text:
            return self.__starttag_text
        return super(HTMLExtractorWithAttribs, self).get_starttag_text()


class CodeBlockClassCheck(Postprocessor):
    """Remove any classes from code blocks that are not language-* classes"""

    def __init__(self, md):
        self._md = md

    def run(self, text: str) -> str:
        print(text)
        root = etree.fromstring(text)
        for code_block in root.findall('.//code'):
            print('Found codee block')
            print(dir(code_block))
            if classes := code_block.attrib.get('class', ''):
                result_classes = ''
                for class_ in classes.split(' '):
                    if class_.startswith('language-'):
                        result_classes.append(class_)
                code_block.attrib['class'] = result_classes
        return self._md.serialise(root)



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
        md.postprocessors.register(CodeBlockClassCheck(md), 'codeblockclasscheck', 21)
        md.treeprocessors.register(LinkAnchorReplacement(md), 'linkanchorreplacement', 1)
        md.treeprocessors.register(ImageSourceCheck(md), 'imagesourcecheck', 19)


def makeExtension(**kwargs):
    """Create instance of markdown extension"""
    return TerraregMarkdownExtension(**kwargs)


def markdown(text, **kwargs):
    """Replicate upstream markdown method to create instance of custom Markdown class"""
    md = CustomMarkdown(**kwargs)
    return md.convert(text)
