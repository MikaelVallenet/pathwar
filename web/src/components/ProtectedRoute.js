import React from "react";
import { useDispatch, useSelector } from "react-redux";
import { Dimmer } from "tabler-react";
import { toast } from "react-toastify";
import { setKeplrSession } from "../actions/userSession";

const ProtectedRoute = ({ component: Component, ...rest }) => {
  const dispatch = useDispatch();
  const userSession = useSelector(state => state.userSession);

  window.onload = async () => {
    try {
      if (!window.keplr) {
        toast.error(
          `Please install Keplr extension to connect your wallet.`
        );
      } else {
        await window.keplr.enable("teritori-1");
        const offlineSigner = window.getOfflineSigner("teritori-1");
        const accounts = await offlineSigner.getAccounts();
        const address = accounts[0].address;
        console.log(address)
        dispatch(setKeplrSession(true))
        const value = await window.keplr.signArbitrary("teritori-1", address, "test");
        console.log(window.keplr.verifyArbitrary("teritori-1", address, "test", value));
      }
    } catch (e) {
      console.log(e);
    }
  }

  if (userSession.isAuthenticated) {
    return <Component {...rest} />;
  }

  return <Dimmer active loader />;
};

export default ProtectedRoute;
