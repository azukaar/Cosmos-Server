// assets
import { MessageOutlined, GithubOutlined, QuestionOutlined, BugOutlined } from '@ant-design/icons';
import { useTheme } from '@mui/material/styles';

// ==============================|| MENU ITEMS - SAMPLE PAGE & DOCUMENTATION ||============================== //

const DiscordOutlinedIcon = (props) => {
    const theme = useTheme();
    return (
        <img src={
            theme.palette.mode === 'dark' ? DiscordOutlinedWhite : DiscordOutlined} width="16px" alt="Discord" {...props} />
    );
};

const support = {
    id: 'support',
    title: 'menu-items.support',
    type: 'group',
    children: [
        {
            id: 'discord',
            title: 'aseracorp.menu-items.support.helpDiscussion',
            type: 'item',
            url: 'https://github.com/aseracorp/resiOS/discussions',
            icon: MessageOutlined,
            external: true,
            target: true
        },
        {
            id: 'github',
            title: 'menu-items.support.github',
            type: 'item',
            url: 'https://github.com/aseracorp/resiOS',
            icon: GithubOutlined,
            external: true,
            target: true
        },
        {
            id: 'documentation',
            title: 'menu-items.support.docsTitle',
            type: 'item',
            url: 'https://cosmos-cloud.io/doc',
            icon: QuestionOutlined,
            external: true,
            target: true
        },
        {
            id: 'bug',
            title: 'menu-items.support.bugReportTitle',
            type: 'item',
            url: 'https://github.com/aseracorp/resiOS/issues',
            icon: BugOutlined,
            external: true,
            target: true
        }
    ]
};

export default support;
