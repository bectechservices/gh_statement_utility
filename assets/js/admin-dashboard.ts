import Vue from "vue";

interface Data {
  socket: WebSocket | null;
  verified: number;
  failed: number;
  successful: number;
  records: Array<any>;
}

interface Methods {
  socketOnMessage: (event: MessageEvent) => void;
  formatNumber: (number: number) => string;
}

export default new Vue<Data, Methods>({
  el: ".dashboardPage",
  beforeMount() {
    this.socket = new WebSocket(`ws://${window.location.hostname}:8001?id=admin`);
    this.socket.addEventListener("message", this.socketOnMessage);
  },
  data: {
    socket: null,
    records:
      (window as any).VERIFICATION_DATA.map((each: any) => {
        return {
          national_id: each.national_id,
          card_id: each.Client.CardID,
          name: `${each.Client.Forenames} ${each.Client.Surname}`,
          created_at: each.created_at,
        };
      }) || [],
    verified: (window as any).TOTAL_VERIFIED,
    failed: (window as any).FAILED_VERIFICATION,
    successful: (window as any).SUCCESSFUL_VERIFICATION,
  },
  methods: {
    socketOnMessage(event: MessageEvent) {
      const message = JSON.parse(event.data);
      if (message.type == "verification_data") {
        this.verified = parseInt(this.verified as any) + 1;
        const content = JSON.parse(message.content);
        if (content.code == "00" && content.success) {
          this.successful = parseInt(this.successful as any) + 1;
          this.records.unshift({
            national_id: content.data.person.nationalId,
            card_id: content.data.person.cardId,
            name: `${content.data.person.forenames} ${content.data.person.surname}`,
            created_at: content.timestamp,
          });
        } else {
          this.failed = parseInt(this.failed as any) + 1;
        }
      }
    },
    formatNumber(number: number) {
      return new Intl.NumberFormat().format(number);
    },
  },
});
