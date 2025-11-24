/*import Vue from "vue";
import axios from "axios";

interface Data {
  loading: boolean;
  search: {
    accountnumber: string;
    startdate: string;
    enddate: string;
  };
  details: {
    name: string;
    branch: string;
    currency: string;
    totalCredit: string;
    totalDebit: string;
  };
  data: Array<{
    entrydate: string;
    valuedate: string;
    particulars: string;
    withdrawal: string;
    deposit: string;
    balance: string;
  }>;
}
interface Methods {
  dataIsValid: () => boolean;
  validate: () => Promise<void>;
  loadAccountDates: () => Promise<void>;
}

export default new Vue<Data, Methods>({
  el: ".statementPage",
  data: {
    loading: false,
    search: {
      accountnumber: "",
      startdate: "",
      enddate: "",
    },
    details: {
      name: "",
      branch: "",
      currency: "",
      totalCredit: "",
      totalDebit: "",
    },
    data: [],
  },
  methods: {
    dataIsValid() {
      let isValid = true;
      (document.getElementById("accno_err") as any).innerText = "";
      (document.getElementById("start_err") as any).innerText = "";
      (document.getElementById("end_err") as any).innerText = "";

      if (this.search.accountnumber.trim().length == 0) {
        (document.getElementById("accno_err") as any).innerText =
          "Account number is required";
        isValid = false;
      }

      if (this.search.startdate.trim().length == 0) {
        (document.getElementById("start_err") as any).innerText =
          "Start date is required";
        isValid = false;
      }

      if (this.search.enddate.trim().length == 0) {
        (document.getElementById("end_err") as any).innerText =
          "End date is required";
        isValid = false;
      }

      return isValid;
    },
    async validate() {
      if (this.dataIsValid()) {
        this.loading = true;
      }
    },
    async loadAccountDates() {
      this.loading = true;
      this.search.startdate = "";
      this.search.enddate = "";

      try {
        const response = await axios.post("/account-dates", {
          accountnumber: this.search.accountnumber.trim(),
        });
        const dates = response.data;
        (this.$refs.startDate as HTMLInputElement).setAttribute(
          "min",
          dates.from
        );
        (this.$refs.startDate as HTMLInputElement).setAttribute(
          "max",
          dates.to
        );
        (this.$refs.endDate as HTMLInputElement).setAttribute(
          "min",
          dates.from
        );
        (this.$refs.endDate as HTMLInputElement).setAttribute("max", dates.to);
      } catch (error) {
        alert((error as any).message);
      } finally {
        this.loading = false;
      }
    },
  },
});*/