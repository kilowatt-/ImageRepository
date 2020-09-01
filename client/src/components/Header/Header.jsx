import React, {useEffect} from "react";
import AppBar from "@material-ui/core/AppBar";
import {Button, Typography} from "@material-ui/core";
import Toolbar from "@material-ui/core/Toolbar";
import {makeStyles} from "@material-ui/core/styles";
import {Link} from "react-router-dom";
import {useUserContext} from "../../context/UserContext";
import Cookies from "universal-cookie";

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
    const cookies = new Cookies();


    useEffect(() => {
        if (!user.name) {
            const cookieUser = cookies.get('userinfo');
            if (cookieUser) {
                dispatch({type: 'setUser', user: cookieUser});
            }
        }
    }, [cookies, dispatch, user]);

    const classes = useStyles();

    const handleLogout = () => {
        cookies.remove('userinfo');
        cookies.remove('logintoken');
        dispatch({type: 'reset'});
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