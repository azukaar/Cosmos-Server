import PropTypes from 'prop-types';

// third-party
import { SyntaxHighlighter } from 'react-syntax-highlighter';
import { a11yDark } from 'react-syntax-highlighter/dist/esm/styles/hljs';
import js from 'react-syntax-highlighter/dist/esm/languages/hljs/javascript';

// ==============================|| CODE HIGHLIGHTER ||============================== //

SyntaxHighlighter.registerLanguage('javascript', js);

export default function SyntaxHighlight({ children, ...others }) {
    return (
        <SyntaxHighlighter language="javascript" showLineNumbers style={a11yDark} {...others}>
            {children}
        </SyntaxHighlighter>
    );
}

SyntaxHighlight.propTypes = {
    children: PropTypes.node
};
