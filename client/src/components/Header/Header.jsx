import React from "react";
import AppBar from "@material-ui/core/AppBar";
import {Button, Typography} from "@material-ui/core";
import Toolbar from "@material-ui/core/Toolbar";
import {makeStyles} from "@material-ui/core/styles";
import {Link} from "react-router-dom";

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
                    <Button color={"inherit"} component={Link} to={'/login'}>Login</Button>
                    <Button color={"inherit"} component={Link} to={'/signup'}>Sign up</Button>
                </Toolbar>
            </AppBar>
        </div>
    )
}

export default Header;