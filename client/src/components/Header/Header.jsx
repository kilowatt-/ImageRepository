import React, {useEffect} from "react";
import AppBar from "@material-ui/core/AppBar";
import {Button, Typography} from "@material-ui/core";
import Toolbar from "@material-ui/core/Toolbar";
import {makeStyles} from "@material-ui/core/styles";
import {Link} from "react-router-dom";
import {useUserContext} from "../../context/UserContext";
import {getUserFromCookie, resetCookie} from "../../utils/getUserFromCookie";

const useStyles = makeStyles(() => ({
    root: {
        flexGrow: 1,
    },
    title: {
        flexGrow: 1,
    },
}));

const Header = () => {
    const [user, dispatch] = useUserContext();

    useEffect(() => {
        if (!user.name) {
            getUserFromCookie(dispatch);
        }
    }, [dispatch, user]);

    const classes = useStyles();

    const handleLogout = () => {
        resetCookie(dispatch);
    }

    return (
        <div className={classes.root}>
            <AppBar position={"static"}>
                <Toolbar>
                    <Typography variant={"h5"} className={classes.title}>Outstagram</Typography>

                    {!user.name ? (
                        <>
                        <Button color={"inherit"} component={Link} to={'/login'}>Login</Button>
                        <Button color={"inherit"} component={Link} to={'/signup'}>Sign up</Button>
                        </>
                        ) :
                    <>
                        <span>Hi, {user.name}!</span>
                        <Button color={"inherit"} onClick={() => handleLogout()}>Log out</Button>
                    </>}
                </Toolbar>
            </AppBar>
        </div>
    )
}

export default Header;