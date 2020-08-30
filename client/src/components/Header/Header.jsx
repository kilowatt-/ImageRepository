import React from "react";
import AppBar from "@material-ui/core/AppBar";
import {Button, Typography} from "@material-ui/core";
import Toolbar from "@material-ui/core/Toolbar";
import {makeStyles} from "@material-ui/core/styles";

const useStyles = makeStyles(() => ({
    root: {
        flexGrow: 1,
    },
    title: {
        flexGrow: 1,
    },
}));

const Header = () => {
    const classes = useStyles();

    return (
        <div className={classes.root}>
            <AppBar position={"static"}>
                <Toolbar>
                    <Typography variant={"h5"} className={classes.title}>Image Repository</Typography>
                    <Button color={"inherit"} href={"/login"}>Login</Button>
                    <Button color={"inherit"} href={"/signup"}>Sign up</Button>
                </Toolbar>
            </AppBar>
        </div>
    )
}

export default Header;