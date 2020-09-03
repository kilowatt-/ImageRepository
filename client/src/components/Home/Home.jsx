import {Button, Container, Typography} from "@material-ui/core";
import Header from "../Header/Header";
import React, {useState} from 'react';
import {useUserContext} from "../../context/UserContext";
import Grid from "@material-ui/core/Grid";
import UploadModal from "./UploadModal";
import Snackbar from "@material-ui/core/Snackbar";
import Alert from "@material-ui/lab/Alert";

const Home = () => {
    const [user] = useUserContext();
    const [uploadModalVisible, setUploadModalVisible] = useState(false);
    const [uploadSuccessSnackbarOpen, setUploadSuccessSnackBarOpen] = useState(false);

    const handleCloseModal = (success) => {
        setUploadModalVisible(false);

        if (success) {
            setUploadSuccessSnackBarOpen(true);
        }
    }

    const handleCloseSnackbar = (event, reason) => {
        if (reason === 'clickaway') return;

        setUploadSuccessSnackBarOpen(false);
    }

    return (
        <div>
            <Header />
            <Container>
                <Snackbar open={uploadSuccessSnackbarOpen} autoHideDuration={6000} onClose={handleCloseSnackbar}>
                    <Alert onClose={handleCloseSnackbar} variant="filled" severity="success">
                        Image uploaded successfully
                    </Alert>
                </Snackbar>
                <Typography variant={"h4"}>Your feed</Typography>
                <Grid container>
                    <UploadModal open={uploadModalVisible} handleClose={handleCloseModal} />
                    {user.name ?  (<Button variant={"contained"} onClick={() => setUploadModalVisible(true)}>+ Upload New Image</Button>) : null}
                    <Grid container spacing={2} justify={"center"}>
                        <Grid item></Grid>
                        <Grid item> h2</Grid>
                        <Grid item> h3</Grid>
                    </Grid>
                </Grid>
            </Container>
        </div>
    )
}

export default Home;