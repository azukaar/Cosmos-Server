import { LeftOutlined } from "@ant-design/icons";
import { IconButton } from "@mui/material";
import { useNavigate } from "react-router";

function Back() {
	const navigate = useNavigate();
	const goBack = () => {
		navigate(-1);
	}
	return <IconButton onClick={goBack}>
		<LeftOutlined />
	</IconButton>	
;
}

export default Back;