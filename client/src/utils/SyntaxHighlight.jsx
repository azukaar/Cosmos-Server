import PropTypes from 'prop-types';

// third-party
import SyntaxHighlighter from 'react-syntax-highlighter';
import { a11yDark } from 'react-syntax-highlighter/dist/esm/styles/hljs';

// ==============================|| CODE HIGHLIGHTER ||============================== //

export default function SyntaxHighlight({ children, ...others }) {
    return (
        <SyntaxHighlighter language="javacript" showLineNumbers style={a11yDark} {...others}>
            {children}
        </SyntaxHighlighter>
    );
}

SyntaxHighlight.propTypes = {
    children: PropTypes.node
};
