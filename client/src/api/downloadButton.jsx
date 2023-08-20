import { Button } from "@mui/material";

export const DownloadFile = ({ filename, content, label }) => {
  const downloadFile = () => {
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
      <Button onClick={downloadFile}>
          {label}
      </Button>
  );
}