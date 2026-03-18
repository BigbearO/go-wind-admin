import {IUser} from "@/models/types";

/**
 * @see https://umijs.org/docs/max/access#access
 * */
export default function access(
  initialState: { currentUser?: IUser } | undefined,
) {
  const {currentUser} = initialState ?? {};
  return {
    canAdmin: currentUser,
  };
}
