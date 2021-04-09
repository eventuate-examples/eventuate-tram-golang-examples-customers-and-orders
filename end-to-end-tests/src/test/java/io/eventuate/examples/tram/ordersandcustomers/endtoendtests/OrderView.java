package io.eventuate.examples.tram.ordersandcustomers.endtoendtests;

public class OrderView {

  private Long id;

  private OrderState state;
  private Money orderTotal;


  public OrderView() {
  }

  public OrderView(Long id, Money orderTotal) {
    this.id = id;
    this.orderTotal = orderTotal;
    this.state = OrderState.PENDING;
  }

  public Money getOrderTotal() {
    return orderTotal;
  }

  public OrderState getState() {
    return state;
  }
}
