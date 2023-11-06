import { ArrowDownOutlined } from "@ant-design/icons";
import { Button } from "@mui/material";
import ResponsiveButton from "../components/responseiveButton";

export const DownloadFile = ({ filename, content, contentGetter, label, style }) => {
    const downloadFile = async () => {
        // Get the content
        if (contentGetter) {
            try {
                content = await contentGetter();
                if(typeof content !== "string") {
                    content = JSON.stringify(content, null, 2);
                }
            } catch (e) {
                console.error(e);
                return;
            }
        }

        // Create a blob with the content
        const blob = new Blob([content], { type: "text/plain;charset=utf-8" });

        // Create a link element
        const link = document.createElement("a");
        link.href = URL.createObjectURL(blob);
        link.download = filename;

        // Append the link to the document (needed for Firefox)
        document.body.appendChild(link);

        // Simulate a click to start the download
        link.click();

        // Cleanup the DOM by removing the link element
        document.body.removeChild(link);
    }

    return (
        <ResponsiveButton
            color="primary"
            onClick={downloadFile}
            style={style}
            variant={"outlined"}
            startIcon={<ArrowDownOutlined />}
        >
            {label}
        </ResponsiveButton>
    );
}