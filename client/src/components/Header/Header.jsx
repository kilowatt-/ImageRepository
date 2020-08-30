import React from "react";
import AppBar from "@material-ui/core/AppBar";
import {Button, Typography} from "@material-ui/core";
import Toolbar from "@material-ui/core/Toolbar";


const Header = () => {
    return (
        <AppBar position={"static"}>
            <Toolbar>
            <Typography variant={"h5"}>Image Repository</Typography>
            <Button>Login</Button>
            </Toolbar>
        </AppBar>
    )
}

export default Header;